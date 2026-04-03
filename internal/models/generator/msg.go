package generator

type GenerateCanceledMsg struct{}

type GenerateSubmitMsg struct {
	Dir    string
	Name   string
	Width  int
	Height int
}

type MazeGeneratedMsg struct {
	Path string
}

type MazeGenerateErrMsg struct {
	Path string
	Err  error
}
