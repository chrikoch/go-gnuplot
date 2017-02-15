package gnuplot

import (
	"os/exec"
	"time"

	"io/ioutil"
	"fmt"
	"os"
)

type Plotter struct {
	data        []TimeDataPoint
	gnuplot_cmd string
}

type TimeDataPoint struct {
	X time.Time
	Y int
}

func NewPlotter() *Plotter {
	var p Plotter
	p.findGnuplotInPath()
	if len(p.gnuplot_cmd) == 0 {
		return nil
	}
	return &p
}

func (p *Plotter) AddTimeDataPoint(d TimeDataPoint) {
	p.data = append(p.data, d)
}

func (p *Plotter) Plot() (image []byte, err error) {
	//get tmp file for data
	datafile, err := ioutil.TempFile("", "plotter")
	if err != nil {
		return image, err
	}
	defer os.Remove(datafile.Name())

	//write data file
	for _, item := range p.data {
		s := fmt.Sprintf("%v %v\n", item.X.Unix(), item.Y)
		_, err = datafile.WriteString(s)
		if err != nil {
			return image, err
		}
	}
	datafile.Close()

	//get tmp file for image
	imagefile, err := ioutil.TempFile("", "plotter")
	if err != nil {
		return image, err
	}
	defer os.Remove(imagefile.Name())
	//close because gnuplot should write it
	imagefile.Close()

	//get tmp file for command
	commandfile, err := ioutil.TempFile("", "plotter")
	if err != nil {
		return image, err
	}
	defer os.Remove(commandfile.Name())

	//write commands to file
	commandfile.WriteString("set terminal png;\n")
	commandfile.WriteString("set output '" + imagefile.Name() + "';\n")
	commandfile.WriteString("set xdata time;\n")
	commandfile.WriteString("set timefmt '%s';\n")
	commandfile.WriteString("plot '" + datafile.Name() + "' using 1:2 title 'abc';\n")
	commandfile.Close()

	err = exec.Command(p.gnuplot_cmd, commandfile.Name()).Run()
	if err != nil {
		return image, err
	}

	image, err = ioutil.ReadFile(imagefile.Name())
	if err != nil {
		return image, err
	}

	return image, nil
}

func (p*Plotter) findGnuplotInPath() {
	p.gnuplot_cmd = ""
	s, err := exec.LookPath("gnuplot")
	if err == nil {
		p.gnuplot_cmd = s
	}
}
