package llm

import llama "github.com/go-skynet/go-llama.cpp"

type Agent struct {
	Model   *string
	Verbose bool
}

func NewAgent(modelPath string, verbose bool) (*Agent, error) {
	model, err := llama.NewModel(modelPath)
	if err != nil {
		return nil, err
	}

	chat, err := llama.NewChat(model)
	if err != nil {
		return nil, err
	}

	return &Agent{
		Model:   model,
		Chat:    chat,
		Verbose: verbose,
	}, nil
}
