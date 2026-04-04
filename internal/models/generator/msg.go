package generator

type GenerateCanceledMsg struct{}

type GenerateSubmitMsg struct {
	Dir    string
	Name   string
	Height int
	Width  int
}

type MazeGeneratedMsg struct {
	Path string
}

type MazeGenerateErrMsg struct {
	Path string
	Err  error
}
