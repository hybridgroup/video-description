package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hybridgroup/yzma/pkg/llama"
	"github.com/hybridgroup/yzma/pkg/mtmd"
)

var libPath = os.Getenv("YZMA_LIB")

var (
	caption string
	tone    string
	humor   string
)

func startCaptions(modelFile, projectorFile, prompt string) {
	if err := llama.Load(libPath); err != nil {
		fmt.Println("unable to load library", err.Error())
		os.Exit(1)
	}
	if err := mtmd.Load(libPath); err != nil {
		fmt.Println("unable to load library", err.Error())
		os.Exit(1)
	}

	if !*verbose {
		llama.LogSet(llama.LogSilent())
	}

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
