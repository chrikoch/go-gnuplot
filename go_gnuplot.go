//Package gnuplot can plot images. Currently only time-based x-Data is supported.
//gnuplot has to be installed!
package gnuplot

import (
	"os/exec"
	"time"

	"fmt"
	"io/ioutil"
	"os"
)

//Plotter represents data and methods to generate a PNG-image using gnuplot
//Don't use Plotter directly, see NewPlotter()
type Plotter struct {
	data       []TimeDataPoint
	gnuplotCmd string
	Title      string //the Title of the plotted data
	XTicsCount int    //if >0, number of xtics to be used (not really accurate)
	YSpace     int    //if >0, percent of space to be left above and under the datapoints
}

//A TimeDataPoint represents a single measurement.
type TimeDataPoint struct {
	X time.Time
	Y int
}

//NewPlotter returns an initialized plotter.
//On error (gnuplot not found in path) nil is returned
func NewPlotter() *Plotter {
	var p Plotter
	p.findGnuplotInPath()
	if len(p.gnuplotCmd) == 0 {
		return nil
	}
	return &p
}

//AddTimeDataPoint adds a new measurement.
func (p *Plotter) AddTimeDataPoint(d TimeDataPoint) {
	p.data = append(p.data, d)
}

//Plot generates a png image from the measurements added by AddTimeDataPoint()
//returns image in PNG format or err on error
func (p *Plotter) Plot() (image []byte, err error) {
	//get tmp file for data
	datafile, err := ioutil.TempFile("", "plotter")
	if err != nil {
		return image, err
	}
	defer os.Remove(datafile.Name())

	//write data file
	var firstTime, lastTime time.Time
	var min, max int
	firstItem := true
	for _, item := range p.data {

		//remember first and last datapoint
		if item.X.Before(firstTime) || firstTime.Equal(time.Time{}) {
			firstTime = item.X
		}
		if item.X.After(lastTime) {
			lastTime = item.X
		}

		//remember highest and lowest Y Value
		if min > item.Y || firstItem {
			min = item.Y
		}
		if max < item.Y || firstItem {
			max = item.Y
		}

		firstItem = false

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

	//xtics
	if p.XTicsCount > 0 && lastTime.After(firstTime) {
		//calculate xtics interval
		xinterval := int(lastTime.Sub(firstTime).Seconds() / float64(p.XTicsCount))
		commandfile.WriteString(fmt.Sprintf("set xtics %v;\n", xinterval))
	}

	//yrange
	if p.YSpace > 0 && max > min {
		space := int(float64((max - min)) * (float64(p.YSpace) / 100))
		commandfile.WriteString(fmt.Sprintf("set yrange [%v:%v];\n", min-space, max+space))
	}

	//plot
	plotCmd := fmt.Sprintf("plot '%v' using 1:2", datafile.Name())
	if len(p.Title) > 0 {
		plotCmd = fmt.Sprintf("%v title '%v'", plotCmd, p.Title)
	}
	plotCmd = plotCmd + ";\n"
	commandfile.WriteString(plotCmd)

	commandfile.Close()

	err = exec.Command(p.gnuplotCmd, commandfile.Name()).Run()
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
	p.gnuplotCmd = ""
	s, err := exec.LookPath("gnuplot")
	if err == nil {
		p.gnuplotCmd = s
	}
}
