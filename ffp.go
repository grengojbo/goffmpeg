package goffmpeg

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grengojbo/beego/modules/utils"
)

var (
	dirVideoVod  string = "/opt/vod/"
	path_ffprobe string = "/usr/local/bin/ffprobe"
	dirConvTs    string = "/opt/convts/"
	dirTs        string = "/opt/ts/"
)

type Transcoder struct {
	Name           string   `json:"name"`
	Filename       string   `json:"filename"`
	IsVideo        bool     `json:"isVideo"`
	IsAudio        bool     `json:"isAudio"`
	DirVideoVod    string   `json:"-"`
	DirConvTs      string   `json:"-"`
	DirTs          string   `json:"-"`
	DirTmp         string   `json:"-"`
	Streams        []Stream `json:"streams"`
	Movie          *Movies  `json:"-"`
	Duration       string   `json:"duration"`
	DurationSeonds int      `json:"duration_seconds"`
	CountVideo     int8     `json:"-"`
	CountAudio     int8     `json:"-"`
	BitRate        string   `json:"-"`
	LogLevel       string   `json:"-"`
}

type Stream struct {
	Index     int64  `json:"index"`
	Name      string `json:"name"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  int    `json:"duration"`
	Filename  string `json:"filename"`
	IsVideo   bool   `json:"-"`
	IsAudio   bool   `json:"-"`
	IsConvert bool   `json:"-"`
}

type StreamVideo struct {
	Id                 int64     `json:"-"`
	Index              int64     `json:"index"`
	Name               string    `json:"name"`
	CodecName          string    `json:"codec_name"`
	CodecType          string    `json:"codec_type"`
	Width              int       `json:"width"`
	Height             int       `json:"height"`
	Duration           string    `json:"duration"`
	DisplayAspectRatio string    `json:"display_aspect_ratio"`
	BitRate            string    `json:"bit_rate"`
	TagsVideo          sTagVideo `json:"tags"`
}

type Movies struct {
	TsVideos []sVideo `json:"streams"`
	Format   sFormat  `json:"format"`
}

type sVideo struct {
	Index              int64        `json:"index"`
	CodecName          string       `json:"codec_name"`
	CodecLongName      string       `json:"codec_long_name"`
	Profile            string       `json:"profile"`
	CodecType          string       `json:"codec_type"`
	CodecTimeBase      string       `json:"codec_time_base"`
	CodecTagString     string       `json:"codec_tag_string"`
	CodecTag           string       `json:"codec_tag"`
	Width              int          `json:"width"`
	Height             int          `json:"height"`
	HasBFrames         int8         `json:"has_b_frames"`
	SampleAspectRatio  string       `json:"sample_aspect_ratio"`
	DisplayAspectRatio string       `json:"display_aspect_ratio"`
	PixFmt             string       `json:"pix_fmt"`
	Level              int          `json:"level"`
	RFrameRate         string       `json:"r_frame_rate"`
	AvgFrameRate       string       `json:"avg_frame_rate"`
	timeBase           string       `json:"time_base"`
	StartPts           int64        `json:"start_pts"`
	startTime          string       `json:"start_time"`
	DurationTs         int64        `json:"duration_ts"`
	Duration           string       `json:"duration"`
	BitRate            string       `json:"bit_rate"`
	NbFrames           string       `json:"nb_frames"`
	Disposition        sDisposition `json:"disposition"`
	TagsVideo          sTagVideo    `json:"tags"`
}

type sFormat struct {
	Filename       string     `json:"filename"`
	NbStreams      int8       `json:"nb_streams"`
	BbPrograms     int8       `json:"nb_programs"`
	FormatName     string     `json:"format_name"`
	FormatLongName string     `json:"format_long_name"`
	startTime      string     `json:"start_time"`
	duration       string     `json:"duration"`
	size           string     `json:"size"`
	bitRate        string     `json:"bit_rate"`
	ProbeScore     int        `json:"probe_score"`
	TagsFormat     sTagFormat `json:"tags"`
}

type sDisposition struct {
	DispDefault     int `json:"default"`
	dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	VleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
}

type sTagVideo struct {
	Language    string `json:"language"`
	HandlerName string `json:"handler_name"`
}

type sTagFormat struct {
	major_brand       string `json:"major_brand"`
	minor_version     string `json:"minor_version"`
	compatible_brands string `json:"compatible_brands"`
	encoder           string `json:"encoder"`
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func bash(str string) {
	run("bash", "-c", str)
}

func run(str ...string) {
	cmd := exec.Command(str[0], str[1:]...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Start()
	cmd.Wait()
}

func runFfmpeg(str ...string) (errMess error) {
	cmd := exec.Command("ffmpeg", str...)
	// stdout, _ := cmd.StdoutPipe()
	_, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	// _, err = cmd.StderrPipe()
	checkError(err)
	errBuf := bufio.NewReader(stderr)
	go func() {
		for {
			errLine, errErr := errBuf.ReadBytes('\n')
			if errErr != nil {
				break
			}
			errMess = errors.New(string(errLine))
			// os.Stdout.Write(errLine)
		}
	}()
	// go io.Copy(os.Stdout, stdout)
	// go io.Copy(os.Stderr, stderr)
	cmd.Start()
	defer cmd.Wait()
	return errMess
}

func PeLs(dir, file string) {
	fconv := fmt.Sprintf("%s%s", dir, file)
	// Replace `ls` (and its arguments) with something more interesting
	//cmd := exec.Command("ls", "-l")
	// cmd := exec.Command("/usr/local/bin/ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", "/opt/vod/z.m4v")
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fconv)

	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)

	// Start command
	err = cmd.Start()
	checkError(err)

	// Don't let main() exit before our command has finished running
	defer cmd.Wait() // Doesn't block

	// Non-blockingly echo command output to terminal
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	// I love Go's trivial concurrency :-D
	fmt.Printf("Do other stuff here! No need to wait.\n\n")
}

// SetFileName("name.ext", "/prefix/", "newext")
// Return: /prefix/name.newext
func SetFileName(s string, p string, e string) (res string) {
	_, filename := filepath.Split(s)
	if len(e) > 0 {
		res = strings.TrimSpace(strings.ToLower(strings.Replace(filename, filepath.Ext(filename), e, 1)))
	} else {
		res = strings.TrimSpace(strings.ToLower(strings.Replace(filename, filepath.Ext(filename), "", 1)))
	}
	if len(p) > 0 {
		res = fmt.Sprintf("%s%s", p, res)
	}
	return res
}

func ListsTs(dirVod string) {
	//lsdir := fmt.Sprintf("%s*", dirVod)
	//files, _ := filepath.Glob(lsdir)
	files, _ := ioutil.ReadDir(dirVod)
	for _, f := range files {
		fext := filepath.Ext(f.Name())
		if fext == ".ts" {
			//filename := strings.TrimSpace(strings.ToLower(strings.Replace(f.Name(), fext, "", 1)))
			//fmt.Println(filename)
			fmt.Println(f.Name())
		}
	}
}

func ListsVod(dirVod string) {
	//lsdir := fmt.Sprintf("%s*", dirVod)
	//files, _ := filepath.Glob(lsdir)
	files, _ := ioutil.ReadDir(dirVod)
	for _, f := range files {
		fext := filepath.Ext(f.Name())
		if fext != ".m4v" {
			filename := strings.TrimSpace(strings.ToLower(strings.Replace(f.Name(), fext, "", 1)))
			fmt.Println(filename)
			fmt.Println(f.Name())
		}
	}
}

func ListsMovie(dirVod string) (fname []string) {
	//lsdir := fmt.Sprintf("%s*", dirVod)
	//files, _ := filepath.Glob(lsdir)
	files, _ := ioutil.ReadDir(dirVod)
	for _, f := range files {
		fext := filepath.Ext(f.Name())
		if fext != ".m4v" {
			//filename := strings.TrimSpace(strings.ToLower(strings.Replace(f.Name(), fext, "", 1)))
			//fmt.Println(filename)
			fname = append(fname, f.Name())
			//fmt.Println(f.Name())
		}
	}
	return fname
}

func (this *Transcoder) Init() {
	this.BitRate = "2000k"
	this.LogLevel = "error"
}

func (this *Transcoder) ListCovert() (fname []string) {
	//lsdir := fmt.Sprintf("%s*", dirVod)
	//files, _ := filepath.Glob(lsdir)
	files, _ := ioutil.ReadDir(this.DirConvTs)
	for _, f := range files {
		fext := filepath.Ext(f.Name())
		if fext != ".m4v" {
			//filename := strings.TrimSpace(strings.ToLower(strings.Replace(f.Name(), fext, "", 1)))
			//fmt.Println(filename)
			fname = append(fname, f.Name())
			//fmt.Println(f.Name())
		}
	}
	return fname
}

func (this *Transcoder) FfProbe(dir, filename string) (movie Movies, err error) {
	arg := []string{
		"-v",
		"quiet",
		"-print_format",
		"json",
		"-show_format",
		"-show_streams",
	}
	// arg = append(arg, fmt.Sprintf("%s%s", dir, file))
	arg = append(arg, filepath.Join(filepath.Dir(dir), filename))
	cmd := exec.Command("ffprobe", arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		//log.Fatal("1 ", err)
		return movie, err
	}
	if err := cmd.Start(); err != nil {
		//log.Fatal("2 ", err)
		return movie, err
	}
	//movie := new(Movies)
	if err := json.NewDecoder(stdout).Decode(&movie); err != nil {
		//log.Fatal("3 ", err)
		return movie, err
	}
	if err := cmd.Wait(); err != nil {
		//log.Fatal("4 ", err)
		return movie, err
	}
	//fmt.Printf("filename: %s", movie.Format.Filename)
	return movie, nil
}

// ffmpeg -i in.avi -r 25 -s 1280x720 -c:a copy -c:v libx264 -b:v 2000k -profile:v high -level 4.0 -pix_fmt yuv420p -threads 4 -ss 00:00:00 -t 00:00:20 -async 1 -vsync 1  -y
func (this *Transcoder) BuildVideo(f string, d int, a string) (fout string, err error) {
	firstPass := true
	arg := []string{
		"-i",
		this.Movie.Format.Filename,
		"-r",
		"25",
		"-loglevel",
		this.LogLevel,
		"-c:v",
		"libx264",
		"-b:v",
		this.BitRate,
		"-profile:v",
		"high",
		"-level",
		"4.0",
		"-pix_fmt",
		"yuv420p",
		"-threads",
		"4",
	}

	// if len(a) > 1 {
	// }
	// arg = append(arg, )
	// arg = append(arg, )
	if f == "hd720" {
		arg = append(arg, "-filter:v")
		arg = append(arg, "scale=iw*min(1280/iw\\,720/ih):ih*min(1280/iw\\,720/ih),pad=1280:720:(1280-iw)/2:(720-ih)/2")
	}
	arg = append(arg, "-ss")
	arg = append(arg, "00:00:00")
	arg = append(arg, "-t")
	arg = append(arg, this.Duration)

	if len(a) > 1 {
		fout = filepath.Join(this.DirTmp, fmt.Sprintf("%s_%s_%s.m4v", this.Name, f, this.BitRate))
		arg = append(arg, "-y")
		arg = append(arg, fout)
		firstPass = false
	} else {
		fout = filepath.Join(this.DirVideoVod, fmt.Sprintf("%s_%s_%s.mp4", this.Name, f, this.BitRate))
		arg = append(arg, "-c:a")
		arg = append(arg, "libfaac")
		arg = append(arg, "-b:a")
		arg = append(arg, "128k")
		arg = append(arg, "-async")
		arg = append(arg, "1")
		arg = append(arg, "-vsync")
		arg = append(arg, "1")
		arg = append(arg, "-y")
		arg = append(arg, fout)

	}
	if errRun := runFfmpeg(arg...); errRun != nil {
		return fout, errRun
	}
	if firstPass {
		fmt.Println("---------------------------------------------------")
		fmt.Println("First video", fout, this.BitRate, this.Duration)
	} else {
		fvideo := fout
		fout = filepath.Join(this.DirVideoVod, fmt.Sprintf("%s_%s_%s.mp4", this.Name, f, this.BitRate))
		// ffmpeg -i /opt/convts/in.m4v -i /opt/convts/in.m4a -c:v copy -c:a copy -ss 00:00:00 -t 00:00:20 -async 1 -vsync 1 /opt/vod/out.mp4
		arg2 := []string{
			"-i",
			fvideo,
			"-i",
			a,
			"-loglevel",
			this.LogLevel,
			"-c:v",
			"copy",
			"-c:a",
			"copy",
			"-async",
			"1",
			"-vsync",
			"1",
			"-y",
			fout,
		}
		if errRun := runFfmpeg(arg2...); errRun != nil {
			return fout, errRun
		}
		// fmt.Println("video", fout, this.BitRate, this.Duration)
	}
	return fout, nil
}
func (this *Transcoder) BuildSilence(d int) (res string, err error) {
	ftemp := filepath.Join(this.DirTmp, fmt.Sprintf("%s_silence.ac3", this.Name))
	fout := filepath.Join(this.DirTmp, fmt.Sprintf("%s_silence.m4a", this.Name))
	fmt.Println("silence", fout)
	arg := []string{
		"-f",
		"lavfi",
		"-i",
		fmt.Sprintf("aevalsrc=0:0:0:0:0:0::duration=%d", d),
		"-loglevel",
		this.LogLevel,
		"-y",
		ftemp,
	}
	// ffmpeg -f lavfi -i aevalsrc=0:0:0:0:0:0::duration=1 -y silence.ac3
	if errRun := runFfmpeg(arg...); errRun != nil {
		return fout, errRun
	}
	arg2 := []string{
		"-i",
		ftemp,
		"-c:a",
		"libfaac",
		"-b:a",
		"128k",
		"-loglevel",
		this.LogLevel,
		"-y",
		fout,
	}
	// ffmpeg -i silence.ac3 -acodec libfaac -b:a 128k silence.m4a
	if errRun := runFfmpeg(arg2...); errRun != nil {
		return fout, errRun
	}
	return fout, nil
}

func (this *Transcoder) BuildTs(f string, suf string) (ts string, err error) {
	// fin, err := this.isExitsVod(f)
	// if err != nil {
	// 	return "", err
	// 	// log.Fatal("is not files: ", fin)
	// }
	fout := filepath.Join(this.DirTs, fmt.Sprintf("%s_%s_%s.ts", this.Name, suf, this.BitRate))
	// fout, err := this.isExitsTs(t)
	// if err != nil {
	// 	return "", err
	// 	// log.Fatal("remove: ", fout)
	// }

	arg := []string{
		"-i",
		f,
		"-loglevel",
		this.LogLevel,
		"-c:v",
		"copy",
		"-c:a",
		"copy",
		"-bsf:v",
		"h264_mp4toannexb",
		"-f",
		"mpegts",
		"-y",
		fout,
	}
	if errRun := runFfmpeg(arg...); errRun != nil {
		return fout, errRun
	}
	return fout, nil
}

// d - продолжительность ролика
func (this *Transcoder) ConvMp4(e string, d int) (fout string, err error) {
	var fileAudio string
	if !this.IsAudio {
		// нет аудио добавляем тишину
		// this.Movie.Format.Filename
		if fileAudio, err = this.BuildSilence(d); err != nil {
			return "", err
		}
		fmt.Println("IsAudio", this.IsAudio, d, e)
	}
	if res, err := this.BuildVideo(e, d, fileAudio); err == nil {
		// fmt.Println("IsVideo", this.IsVideo, "AUDIO file", fileAudio, "src", res)
		return res, nil
	} else {
		return "", err
	}
}

// TODO remove
func (this *Transcoder) Conv(t string, fromFile string, name string, e string) error {
	filename := strings.TrimSpace(strings.ToLower(strings.Replace(name, filepath.Ext(name), "."+e, 1)))
	fout := filepath.Join(this.DirTmp, filename)
	switch {
	case e == "mp4":
		fout = filepath.Join(this.DirVideoVod, filename)
	case e == "ts":
		fout = filepath.Join(this.DirTs, filename)
	}
	// fmt.Printf("CONV: %s -> %s\n", fromFile, fout)
	arg := []string{
		"-i",
		fromFile,
		"-c:a",
		"libfaac",
		"-b:a",
		"128k",
		"-c:v",
		"libx264",
		"-profile:v",
		"main",
		"-b:v",
		"2000k",
		"-filter:v",
		"scale=iw*min(1280/iw\\,720/ih):ih*min(1280/iw\\,720/ih),pad=1280:720:(1280-iw)/2:(720-ih)/2",
		"-async",
		"1",
		"-vsync",
		"1",
		fout,
	}
	// ffmpeg -i /opt/convts/Cinque02.avi -ss 00:00:00 -t 00:00:20 -acodec pcm_s32le /opt/convts/Cinque02.wav
	//ffmpeg -f lavfi -i aevalsrc=0:0:0:0:0:0::duration=1 silence.ac3
	// ffmpeg -i /opt/convts/Cinque02.wav -acodec libfaac -b:a 128k /opt/convts/Cinque02.m4a
	// ffmpeg -i /opt/convts/Cinque02.avi -ss 00:00:00 -t 00:00:20  -s hd720 -vcodec libx264 -b:v 3000k -an /opt/convts/Cinque02.m4v
	// ffmpeg -i /opt/convts/Cinque02.m4v -i /opt/convts/Cinque02.m4a -c:v copy -c:a copy -ss 00:00:00 -t 00:00:20 -async 1 -vsync 1 /opt/vod/cinque02.mp4
	//ffmpeg -i /opt/vod/screan_night_food.mp4 -r 25 -s 1280x720 -c:a copy -c:v libx264 -b:v 2000k -profile:v high -level 4.0 -pix_fmt yuv420p -threads 4 -ss 00:00:00 -t 00:00:20 -async 1 -vsync 1  -y /opt/vod/screan_night_food1.mp4
	//ffmpeg -i /opt/vod/screan_night_food1.mp4 -c:v copy -c:a copy -bsf:v h264_mp4toannexb -f mpegts  -y /opt/channel1/screan_night_food.ts

	// fmt.Printf(">>> %s", filename)
	// var varg []string
	// varg = append(varg, "-s")
	// varg = append(varg, "hd720")
	// //varg = append(varg, "")
	cmd := exec.Command("ffmpeg", arg...)
	// cmd := exec.Command("ffmpeg", "-i", fullname, "-acodec", "libfaac", "-b:a", "128k", "-vcodec", "libx264", "-vprofile", "main", "-b:v", "2000k", "-filter:v", "scale=iw*min(1280/iw\\,720/ih):ih*min(1280/iw\\,720/ih),pad=1280:720:(1280-iw)/2:(720-ih)/2", "-async", "1", "-vsync", "1", fout)
	// //cmd := exec.Command("ffmpeg", "-i", fullname, "-s", "hd720", "-acodec", "libfaac", "-b:a", "128k", "-vcodec", "libx264", "-b:v", "2000k", "-async", "1", "-vsync", "1", fout)
	// //ffmpeg -i $DOUT/$FIN -s hd720 -acodec libfaac -b:a 128k -vcodec libx264 -b:v 2000k -async 1 -vsync 1 $DIN/$FOUT.mp4

	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)

	// Start command
	err = cmd.Start()
	checkError(err)

	// Don't let main() exit before our command has finished running
	defer cmd.Wait() // Doesn't block

	// Non-blockingly echo command output to terminal
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	return nil
}

// func (this Transcoder) isExitsVod(filename string) (string, error) {
// 	dir := filepath.Dir(this.DirVideoVod)
// 	fname := filepath.Join(dir, filename)
// 	if _, err := os.Stat(fname); os.IsNotExist(err) {
// 		return fname, err
// 	} else {
// 		return fname, nil
// 	}
// }

// func (this Transcoder) isExitsConv(filename string) (string, error) {
// 	dir := filepath.Dir(this.DirConvTs)
// 	fname := filepath.Join(dir, filename)
// 	if _, err := os.Stat(fname); os.IsNotExist(err) {
// 		return fname, err
// 	} else {
// 		return fname, nil
// 	}
// }

// func (this Transcoder) isExitsTs(filename string) (string, error) {
// 	dir := filepath.Dir(this.DirTs)
// 	fname := filepath.Join(dir, filename)
// 	if _, err := os.Stat(fname); os.IsNotExist(err) {
// 		return fname, err
// 	} else {
// 		return fname, nil
// 	}
// }

func (this Transcoder) SetMovie(s string) (m Movies) {
	this.IsVideo = false
	this.IsAudio = false
	var dir string
	var filename string
	fullname := filepath.Join(this.DirConvTs, s)
	if err := utils.IsExits(s); err == nil {
		tdir, tfilename := filepath.Split(s)
		dir = tdir + "/"
		filename = tfilename
		// fmt.Println("Convert: ", c.Args().First(), " lang: ", c.GlobalString("lang"))
	} else if err := utils.IsExits(fullname); err == nil {
		dir = this.DirConvTs
		filename = s
	}
	if movie, err := this.FfProbe(dir, filename); err != nil {
		log.Fatal(err)
		// fmt.Println("Convert: ", filename)
		// fmt.Println(movie)
	} else {
		m = movie
	}
	return m
}

func (m Movies) SetStream(t *Transcoder) {
	// t.DurationSeonds = 1
	t.Movie = &m
	t.IsVideo = false
	t.IsAudio = false
	t.CountVideo = 0
	t.CountAudio = 0
	t.Name = SetFileName(m.Format.Filename, "", "")
	for _, st := range m.TsVideos {
		// dur := strings.Split(st.Duration, ".")
		durs, dur := utils.DurationToTime(st.Duration)
		if durs > 0 {
			t.DurationSeonds = durs
			t.Duration = dur
		}
		s := Stream{Index: st.Index, CodecName: st.CodecName, CodecType: st.CodecType, Duration: durs}
		if st.CodecType == "video" {
			s.IsVideo = true
			t.IsVideo = true
			t.CountVideo++
			s.Width = st.Width
			s.Height = st.Height
			fmt.Printf("VIDEO index: %d (%dx%d) duration: %d codec: %s\n", s.Index, s.Width, s.Height, s.Duration, s.CodecName)
		} else if st.CodecType == "audio" {
			s.IsAudio = true
			t.IsAudio = true
			t.CountAudio++
			fmt.Printf("AUDIO index: %d duration: %d codec: %s\n", s.Index, s.Duration, s.CodecName)
		}
	}

	fmt.Println("timeDuration: ", t.Duration, " Video: ", t.IsVideo, "(", t.CountVideo, ")", " Audio: ", t.IsAudio)
	fmt.Printf("filepath: %s %s %s\n", t.Movie.Format.Filename, t.Name, SetFileName(t.Movie.Format.Filename, "./", ".ts"))
}

// TODO remove
func (this Transcoder) SetStream(m Movies) {
	this.Movie = &m
	this.IsVideo = false
	this.IsAudio = false
	this.CountVideo = 0
	this.CountAudio = 0
	this.Name = SetFileName(m.Format.Filename, "", "")
	for _, st := range m.TsVideos {
		// dur := strings.Split(st.Duration, ".")
		durs, dur := utils.DurationToTime(st.Duration)
		if durs > 0 {
			this.DurationSeonds = durs
			this.Duration = dur
		}
		s := Stream{Index: st.Index, CodecName: st.CodecName, CodecType: st.CodecType, Duration: durs}
		if st.CodecType == "video" {
			s.IsVideo = true
			this.IsVideo = true
			this.CountVideo++
			s.Width = st.Width
			s.Height = st.Height
			fmt.Printf("VIDEO index: %d (%dx%d) duration: %d codec: %s\n", s.Index, s.Width, s.Height, s.Duration, s.CodecName)
		} else if st.CodecType == "audio" {
			s.IsAudio = true
			this.IsAudio = true
			this.CountAudio++
			fmt.Printf("AUDIO index: %d duration: %d codec: %s\n", s.Index, s.Duration, s.CodecName)
		}
	}

	fmt.Println("timeDuration: ", this.Duration, " Video: ", this.IsVideo, "(", this.CountVideo, ")", " Audio: ", this.IsAudio)
	fmt.Printf("filepath: %s %s %s\n", this.Movie.Format.Filename, this.Name, SetFileName(this.Movie.Format.Filename, "./", ".ts"))
	// var dir string
	// var filename string
	// fullname := filepath.Join(this.DirConvTs, s)
	// if err := utils.IsExits(s); err == nil {
	// 	tdir, tfilename := filepath.Split(s)
	// 	dir = tdir + "/"
	// 	filename = tfilename
	// 	// fmt.Println("Convert: ", c.Args().First(), " lang: ", c.GlobalString("lang"))
	// } else if err := utils.IsExits(fullname); err == nil {
	// 	dir = this.DirConvTs
	// 	filename = s
	// }
	// if movie, err := this.FfProbe(dir, filename); err != nil {
	// 	log.Fatal(err)
	// 	// fmt.Println("Convert: ", filename)
	// 	// fmt.Println(movie)
	// } else {
	// 	m = movie
	// }
	// return m
	// if err := utils.IsExits(m.Format.Filename); err == nil {
	// if fullname, err := this.isExitsConv(filename); err != nil {
	// log.Fatal("no such file or directory: ", fullname)
	// } else {
	// 	fn := strings.TrimSpace(strings.ToLower(strings.Replace(filename, filepath.Ext(filename), ".mp4", 1)))
	// 	fts := strings.TrimSpace(strings.ToLower(strings.Replace(filename, filepath.Ext(filename), ".ts", 1)))
	// 	if movie, err := this.FfProbe(this.DirConvTs, filename); err != nil {
	// 		log.Fatal(err)
	// 	} else {
	// 		this.Movie = movie
	// 		//fmt.Printf(">>> %v", movie)
	// 		for _, st := range movie.TsVideos {
	// 			if st.CodecType == "video" {
	// 				this.IsVideo = true
	// 				fmt.Printf("file: %s(%dx%d) duration: %s  %s\n", fullname, st.Width, st.Height, st.Duration, st.CodecName)

	// 				_ = this.Conv("720p", st.Width, st.Height, fn, fullname)
	// 				_ = this.BuildTs(fn, fts)
	// 			} else if st.CodecType == "audio" {
	// 				this.IsAudio = true
	// 				fmt.Printf("audio: %s\n", st.CodecName)
	// 			}
	// 		}
	// 	}
	// }
}

func (this Movies) Show() {
	// isVideo := false
	// isAudio := false
	// vDuration := 0
	// aDuration := 0
	fmt.Println("file: ", this.Format.Filename)
	for _, st := range this.TsVideos {
		// dur := strings.Split(st.Duration, ".")
		d, timeDuration := utils.DurationToTime(st.Duration)
		fmt.Println("timeDuration: ", timeDuration)
		if st.CodecType == "video" {
			// isVideo = true
			fmt.Printf("VIDEO index: %d (%dx%d) duration: %d codec: %s\n", st.Index, st.Width, st.Height, d, st.CodecName)
		} else if st.CodecType == "audio" {
			// isAudio = true
			fmt.Printf("AUDIO index: %d duration: %d codec: %s\n", st.Index, d, st.CodecName)
		}
	}
}

//func (this Transcoder) setMovie(movie Movies) {
//}

// func ffConvertOne() {
// 	tr := Transcoder{DirVideoVod: "/opt/vod/", DirConvTs: "/opt/convts/", DirTs: "/opt/channel1/"}
// 	testFile := "de_glaimond.mov"
// 	tr.setMovie(testFile)
// }

// func ffProbeTest() {
// 	tr := Transcoder{DirVideoVod: "/opt/vod/", DirConvTs: "/opt/convts/", DirTs: "/opt/channel1/"}
// 	//fmt.Printf("vod: %s\n", tr.dirVideoVod)
// 	testFile := "aabb_inside.avi"
// 	if err := isExits(testFile); err != nil {
// 		fmt.Printf("no such file or directory: %s\n", testFile)
// 		if fullname, err := tr.isExitsConv(testFile); err != nil {
// 			log.Fatal("no such file or directory: ", fullname)
// 		} else {
// 			if _, err := ffProbe(fullname); err != nil {
// 				log.Fatal(err)
// 				//} else {
// 				//	tr.setMovie(movie)
// 			}
// 		}
// 	}
// }

// func main() {
// 	//ListsTs("/opt/channel1/")
// 	resConv := ListsMovie("/opt/convts/")
// 	resTs := ListsMovie("/opt/channel1/")
// 	resVod := ListsMovie("/opt/vod/")
// 	fmt.Println(resConv)
// 	fmt.Println(resTs)
// 	fmt.Println(resVod)
// 	//ListsVod(dirVideoVod)
// 	//ffConvertOne()
// 	//ffProbeTest()
// 	//eLs()
// }
