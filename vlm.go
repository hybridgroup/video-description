package main

import (
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/hybridgroup/yzma/pkg/llama"
	"github.com/hybridgroup/yzma/pkg/mtmd"
)

var libPath = os.Getenv("YZMA_LIB")

var (
	caption string
	tone    string
	humor   string
)

func startVLM(modelFile, projectorFile, prompt string) {
	if err := llama.Load(libPath); err != nil {
		fmt.Println("unable to load library", err.Error())
		os.Exit(1)
	}
	if err := mtmd.Load(libPath); err != nil {
		fmt.Println("unable to load library", err.Error())
		os.Exit(1)
	}

	llama.LogSet(llama.LogSilent())

	llama.Init()
	defer llama.BackendFree()

	vlm := NewVLM(modelFile, projectorFile)
	if err := vlm.Init(); err != nil {
		fmt.Println("unable to initialize VLM:", err)
		os.Exit(1)
	}
	defer vlm.Close()

	for {
		caption = nextCaption(vlm, prompt)
		fmt.Println("Caption:", caption)

		time.Sleep(3 * time.Second)
	}
}

func nextCaption(vlm *VLM, prompt string) string {
	newPrompt := prompt + promptStyle() + mtmd.DefaultMarker()
	fmt.Println(newPrompt)

	messages := []llama.ChatMessage{llama.NewChatMessage("user", newPrompt)}
	input := mtmd.NewInputText(vlm.ChatTemplate(messages, true), true, true)

	bitmap, err := matToBitmap(img)
	if err != nil {
		fmt.Println("Error converting image to bitmap:", err)
		return ""
	}
	defer mtmd.BitmapFree(bitmap)

	output := mtmd.InputChunksInit()
	defer mtmd.InputChunksFree(output)

	if err := vlm.Tokenize(input, bitmap, output); err != nil {
		fmt.Println("Error tokenizing input:", err)
		return ""
	}

	results, err := vlm.Results(output)
	if err != nil {
		fmt.Println("Error obtaining VLM results:", err)
		return ""
	}

	return results
}

const keepShort = " Keep the response to 30 words or less."

func promptStyle() string {
	switch {
	case tone == "" && humor == "":
		return keepShort
	case tone != "" && humor != "":
		return " Be both " + tone + " and " + humor + " in your response." + keepShort
	case tone == "" && humor != "":
		return " Be somewhat " + humor + " in your response." + keepShort
	case tone != "" && humor == "":
		return " Be somewhat " + tone + " in your response." + keepShort
	}

	return keepShort
}

// VLM is a Vision Language Model (VLM).
type VLM struct {
	TextModelFilename      string
	ProjectorModelFilename string

	TextModel        llama.Model
	Sampler          llama.Sampler
	ModelContext     llama.Context
	ProjectorContext mtmd.Context

	template string
}

// NewVLM creates a new VLM.
func NewVLM(model, projector string) *VLM {
	return &VLM{
		TextModelFilename:      model,
		ProjectorModelFilename: projector,
	}
}

// Close closes the VLM.
func (m *VLM) Close() {
	if m.ProjectorContext != 0 {
		mtmd.Free(m.ProjectorContext)

	}

	if m.ModelContext != 0 {
		llama.Free(m.ModelContext)
	}
}

func (m *VLM) Init() error {
	m.TextModel = llama.ModelLoadFromFile(m.TextModelFilename, llama.ModelDefaultParams())

	ctxParams := llama.ContextDefaultParams()
	ctxParams.NCtx = 4096
	ctxParams.NBatch = 2048

	m.ModelContext = llama.InitFromModel(m.TextModel, ctxParams)

	m.template = llama.ModelChatTemplate(m.TextModel, "")

	m.Sampler = llama.NewSampler(m.TextModel, llama.DefaultSamplers)

	mtmdCtxParams := mtmd.ContextParamsDefault()
	mtmdCtxParams.Verbosity = llama.LogLevelContinue
	m.ProjectorContext = mtmd.InitFromFile(m.ProjectorModelFilename, m.TextModel, mtmdCtxParams)

	return nil
}

func (m *VLM) ChatTemplate(messages []llama.ChatMessage, add bool) string {
	buf := make([]byte, 1024)
	len := llama.ChatApplyTemplate(m.template, messages, add, buf)
	result := string(buf[:len])

	return result
}

func (m *VLM) Tokenize(input *mtmd.InputText, bitmap mtmd.Bitmap, output mtmd.InputChunks) (err error) {
	if res := mtmd.Tokenize(m.ProjectorContext, output, input, []mtmd.Bitmap{bitmap}); res != 0 {
		err = fmt.Errorf("unable to tokenize: %d", res)
	}
	return
}

func (m *VLM) Results(output mtmd.InputChunks) (string, error) {
	var n llama.Pos
	nBatch := llama.NBatch(m.ModelContext)

	if res := mtmd.HelperEvalChunks(m.ProjectorContext, m.ModelContext, output, 1, 0, int32(nBatch), true, &n); res != 0 {
		return "", errors.New("unable to evaluate chunks")
	}

	var sz int32 = 1
	batch := llama.BatchInit(1, 0, 1)
	defer llama.BatchFree(batch)

	batch.NSeqId = &sz
	batch.NTokens = 1
	seqs := unsafe.SliceData([]llama.SeqId{0})
	batch.SeqId = &seqs

	vocab := llama.ModelGetVocab(m.TextModel)
	results := ""

	for i := 0; i < int(nBatch); i++ {
		token := llama.SamplerSample(m.Sampler, m.ModelContext, -1)

		if llama.VocabIsEOG(vocab, token) {
			break
		}

		buf := make([]byte, 128)
		len := llama.TokenToPiece(vocab, token, buf, 0, true)
		results += string(buf[:len])

		batch.Token = &token
		batch.Pos = &n

		llama.Decode(m.ModelContext, batch)
		n++
	}

	m.Clear()

	return results, nil
}

// Clear clears the context memory.
func (m *VLM) Clear() {
	llama.Synchronize(m.ModelContext)
	llama.MemoryClear(llama.GetMemory(m.ModelContext), true)
}
