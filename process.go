package minotar

import (
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/png"
	//"image/color"
	"errors"
	"io"
)

func cropImage(i image.Image, d image.Rectangle) (image.Image, error) {
	bounds := i.Bounds()
	if bounds.Min.X > d.Min.X || bounds.Min.Y > d.Min.Y || bounds.Max.X < d.Max.X || bounds.Max.Y < d.Max.Y {
		return nil, errors.New("Bounds invalid for crop")
	}

	dims := d.Size()
	outIm := image.NewRGBA(image.Rect(0, 0, dims.X, dims.Y))
	for x := 0; x < dims.X; x++ {
		for y := 0; y < dims.Y; y++ {
			outIm.Set(x, y, i.At(d.Min.X+x, d.Min.Y+y))
		}
	}
	return outIm, nil
}

type Skin struct {
	Image image.Image
}

func (i Skin) Head() (image.Image, error) {
	return cropImage(i.Image, image.Rect(8, 8, 16, 16))
}

func (i Skin) Helm() (image.Image, error) {
	headImg, err := i.Head()
	if err != nil {
		return nil, err
	}

	headImgRGBA := headImg.(*image.RGBA)

	helmImg, err := cropImage(i.Image, image.Rect(40, 8, 48, 16))
	if err != nil {
		return nil, err
	}

	sr := helmImg.Bounds()
	draw.Draw(headImgRGBA, sr, helmImg, sr.Min, draw.Over)

	return headImg, nil
}

func DecodeSkin(r io.Reader) (Skin, error) {
	skinImg, _, err := image.Decode(r)
	if err != nil {
		return Skin{}, err
	}
	return Skin{
		Image: skinImg,
	}, err
}

func WritePNG(w io.Writer, i image.Image) error {
	return png.Encode(w, i)
}

func Resize(width, height uint, img image.Image) image.Image {
	return resize.Resize(width, height, img, resize.NearestNeighbor)
}
