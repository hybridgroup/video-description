# Captions With Attitude

![Captions With Attitude](./images/vlm-caption.png)

"Captions With Attitude" is a Go application that uses a Vision Language Model (VLM) to show live captions from your webcam in your browser all running entirely on your local machine!

It uses [yzma](https://github.com/hybridgroup/yzma) to perform local inference using [`llama.cpp`](https://github.com/ggml-org/llama.cpp) and [GoCV](https://github.com/hybridgroup/gocv) for the video processing.

## Installation

### yzma

You must install yzma and llama.cpp to run this program.

See https://github.com/hybridgroup/yzma/blob/main/INSTALL.md

### GoCV

You must also install OpenCV and GoCV, which unlike yzma requires CGo.

See https://gocv.io/getting-started/

Although yzma does not use CGo, yzma can co-exist in Go applications that use CGo.

### Models

You will need a Vision Language Model (VLM). Download the model and projector files from Hugging Face in `.gguf` format.

For example, you can use the Qwen3-VL-2B-Instruct model.

https://huggingface.co/ggml-org/Qwen3-VL-2B-Instruct-GGUF

## Building

```shell
go build .
```

## Running

```shell
./captions-with-attitude 0 localhost:8080 ~/models/Qwen3-VL-2B-Instruct-Q8_0.gguf ~/models/mmproj-Qwen3-VL-2B-Instruct-Q8_0.gguf "Give a very brief description of what is going on."
```

Now open your web browser pointed to http://localhost:8080/
