package main

import (
	"bytes"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"image"
	"log"
)

func getPalette(ctx *fasthttp.RequestCtx) {
	imgUrl := ctx.QueryArgs().Peek("img")
	if len(imgUrl) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		_, _ = ctx.WriteString("Not enough params")
		return
	}

	req := fasthttp.AcquireRequest()
	req.SetRequestURIBytes(imgUrl)
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		_, _ = ctx.WriteString(err.Error())
		return
	}

	buffer := resp.Body()
	statusCode := resp.StatusCode()

	if statusCode != fasthttp.StatusOK {
		ctx.SetStatusCode(statusCode)
		_, _ = ctx.Write(buffer)
		return
	}

	img, _, err := image.Decode(bytes.NewReader(buffer))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, _ = ctx.WriteString(err.Error())
		return
	}

	palette := GetPalette(img, 6)
	js, err := json.Marshal(palette)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		_, _ = ctx.WriteString(err.Error())
		return
	}
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Cache-Control", "s-maxage=3600, stale-while-revalidate")
	_, _ = ctx.Write(js)
}

func main() {

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			getPalette(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	// pass plain function to fasthttp
	if err := fasthttp.ListenAndServe(":8081", requestHandler); err != nil {
		log.Fatal(err)
	}
}
