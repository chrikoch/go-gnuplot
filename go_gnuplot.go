package gnuplot

import (
	"os/exec"
	"time"

	"fmt"
	"io/ioutil"
	"os"
)

type Plotter struct {
	data        []TimeDataPoint
	gnuplot_cmd string
	Title       string //the Title of the plotted data
	XTicsCount  int    //if >0, number of xtics to be used (not really accurate)
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
	var firstTime, lastTime time.Time
	for _, item := range p.data {

		//remember first and last datapoint
		if item.X.Before(firstTime) || firstTime.Equal(time.Time{}) {
			firstTime = item.X
		}
		if item.X.After(lastTime) {
			lastTime = item.X
		}

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
	plotCmd := fmt.Sprintf("plot '%v' using 1:2", datafile.Name())
	if len(p.Title) > 0 {
		plotCmd = fmt.Sprintf("%v title '%v'", plotCmd, p.Title)
	}
	plotCmd = plotCmd + ";\n"
	commandfile.WriteString(plotCmd)

	if p.XTicsCount > 0 && lastTime.After(firstTime) {

		//calculate xtics interval
		xinterval := int(lastTime.Sub(firstTime).Seconds() / float64(p.XTicsCount))
		commandfile.WriteString(fmt.Sprintf("set xtics %v;\n", xinterval))
	}

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

func (p *Plotter) findGnuplotInPath() {
	p.gnuplot_cmd = ""
	s, err := exec.LookPath("gnuplot")
	if err == nil {
		p.gnuplot_cmd = s
	}
}
