package main

import (
	"encoding/json"
	"github.com/kataras/iris"
	"github.com/kataras/iris/cache"
	"github.com/kataras/iris/middleware/recover"
	"image"
	"net/http"
	"time"
)

func main() {
	app := iris.Default()

	app.Use(recover.New())

	caching := cache.Handler(1 * time.Hour)

	app.Get("/", caching, func(ctx iris.Context) {
		imgUrl := ctx.URLParam("img")
		if imgUrl == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			_, _ = ctx.WriteString("Not enough params")
			return
		}
		res, err := http.Get(imgUrl)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_, _ = ctx.WriteString(err.Error())
			return
		}
		defer res.Body.Close()
		img, _, err := image.Decode(res.Body)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_, _ = ctx.WriteString(err.Error())
			return
		}
		palette := GetPalette(img, 6)
		js, err := json.Marshal(palette)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_, _ = ctx.WriteString(err.Error())
			return
		}
		ctx.Header("Content-Type", "application/json")
		ctx.Header("Cache-Control", "s-maxage=3600, stale-while-revalidate")
		_, _ = ctx.Write(js)
	})

	_ = app.Run(iris.Addr(":8080"), iris.WithOptimizations)
}
