package domain

type Cell struct {
	RightWall  bool
	BottomWall bool
}

type Board struct {
	Height, Width int
	Cells         [][]Cell
}
