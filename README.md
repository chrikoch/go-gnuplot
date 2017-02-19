<span style="display: inline-block;">
[![API Reference](http://img.shields.io/badge/api-reference-blue.svg)](https://godoc.org/github.com/chrikoch/go-gnuplot)
</span>

# go-gnuplot
Golang wrapper for gnuplot command line <br>
gnuplot has to be installed!

Limitations
-----------
Currently only timebased x-values are supported.

Example
-------

```go
	p := gnuplot.NewPlotter()
	if p == nil {
		return
	}
	point := gnuplot.TimeDataPoint{X: time.Now(), Y: 2}
	p.AddTimeDataPoint(point)
	point = gnuplot.TimeDataPoint{X: time.Now().Add(time.Second * 10), Y: 3}
	p.AddTimeDataPoint(point)
	point = gnuplot.TimeDataPoint{X: time.Now().Add(time.Second * 20), Y: 2}
	p.AddTimeDataPoint(point)

	img, err := p.Plot()
	if err != nil {
		log.Println(err)
		return
	}

	ioutil.WriteFile("/tmp/meinfile.png", img, os.ModePerm)

```
