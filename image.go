package main

type Image struct {
	Dir string
}

func (i *Image) LayerPath() (id string) {
	return i.Dir + "/layer"
}
