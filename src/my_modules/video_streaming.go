package my_modules

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func UploadVideoForStream(c *gin.Context, base_path string, video_title string, inputFile string) error {
	return CreateHLS(inputFile, fmt.Sprintf("%s/multi_bitrate/%s", base_path, video_title), 10)
}

func CreateHLS(inputFile string, outputDir string, segmentDuration int) error {
	// Create the output directory if it does not exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Create the HLS playlist and segment the video using ffmpeg

	var random_string string
	if _rand, r_err := randomBytes(16); r_err != nil {
		return r_err
	} else {
		random_string = hex.EncodeToString(_rand)
	}

	key_file_path := fmt.Sprintf("%s/key.txt", outputDir)
	key_info_file_path := fmt.Sprintf("%s/key_info.txt", outputDir)
	key_info := fmt.Sprintf("/cdn/%s\n%s", key_file_path, key_file_path)
	if err := ioutil.WriteFile(key_file_path, []byte(random_string), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(key_info_file_path, []byte(key_info), 0755); err != nil {
		return err
	}

	// ffmpegCmd := exec.Command(
	// 	"ffmpeg",
	// 	"-i", inputFile,
	// 	"-profile:v", "baseline", // baseline profile is compatible with most devices
	// 	"-level", "3.0",
	// 	"-start_number", "0", // start numbering segments from 0
	// 	"-hls_time", strconv.Itoa(segmentDuration), // duration of each segment in seconds
	// 	"-hls_list_size", "0", // keep all segments in the playlist
	// 	"-hls_key_info_file", key_info_file_path,
	// 	"-f", "hls",
	// 	fmt.Sprintf("%s/playlist.m3u8", outputDir),
	// )

	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-i", inputFile,
		"-filter_complex", `[0:v]split=4[v1][v2][v3][v4]; [v1]copy[v1out]; [v2]scale=w=1280:h=720:force_original_aspect_ratio=decrease[v2out]; [v3]scale=w=640:h=360:force_original_aspect_ratio=decrease[v3out]; [v4]scale=w=360:h=128:force_original_aspect_ratio=decrease[v4out]`,
		"-map", "[v1out]", "-c:v:0", "libx264", "-b:v:0", "10M", "-maxrate:v:0", "10M", "-bufsize:v:0", "15M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "[v2out]", "-c:v:1", "libx264", "-b:v:1", "3M", "-maxrate:v:1", "3M", "-bufsize:v:1", "3M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "[v3out]", "-c:v:2", "libx264", "-b:v:2", "1M", "-maxrate:v:2", "1M", "-bufsize:v:2", "1M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "[v4out]", "-c:v:3", "libx264", "-b:v:3", "0.5M", "-maxrate:v:3", "0.5M", "-bufsize:v:3", "0.5M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
		"-map", "a:0", "-c:a:0", "aac", "-b:a:0", "128k", "-ac", "2",
		"-map", "a:0", "-c:a:1", "aac", "-b:a:1", "96k", "-ac", "2",
		"-map", "a:0", "-c:a:2", "aac", "-b:a:2", "48k", "-ac", "2",
		"-map", "a:0", "-c:a:3", "aac", "-b:a:3", "28k", "-ac", "2",
		"-f", "hls", "-hls_time", "2", "-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments", "-hls_segment_type", "mpegts",
		"-hls_segment_filename", fmt.Sprintf("%s/playlist_%%v/data%%02d.ts", outputDir),
		"-master_pl_name", "playlist.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3",
		fmt.Sprintf("%s/playlist_%%v/manifest.m3u8", outputDir),
	)

	{
		ffprobeCmd := exec.Command(
			"ffprobe",
			"-i", inputFile,
			"-show_streams",
		)
		output, err := ffprobeCmd.CombinedOutput()

		if err != nil {
			return err
		}
		if err == nil && !strings.Contains(string(output), "Stream #0:1") {
			ffmpegCmd = exec.Command(
				"ffmpeg",
				"-i", inputFile,
				"-filter_complex", `[0:v]split=4[v1][v2][v3][v4]; [v1]copy[v1out]; [v2]scale=w=1280:h=720:force_original_aspect_ratio=decrease[v2out]; [v3]scale=w=640:h=360:force_original_aspect_ratio=decrease[v3out]; [v4]scale=w=360:h=128:force_original_aspect_ratio=decrease[v4out]`,
				"-map", "[v1out]", "-c:v:0", "libx264", "-b:v:0", "10M", "-maxrate:v:0", "10M", "-bufsize:v:0", "15M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
				"-map", "[v2out]", "-c:v:1", "libx264", "-b:v:1", "3M", "-maxrate:v:1", "3M", "-bufsize:v:1", "3M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
				"-map", "[v3out]", "-c:v:2", "libx264", "-b:v:2", "1M", "-maxrate:v:2", "1M", "-bufsize:v:2", "1M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
				"-map", "[v4out]", "-c:v:3", "libx264", "-b:v:3", "0.5M", "-maxrate:v:3", "0.5M", "-bufsize:v:3", "0.5M", "-preset", "medium", "-g", "48", "-sc_threshold", "0", "-keyint_min", "48",
				"-f", "hls", "-hls_time", "2", "-hls_playlist_type", "vod",
				"-hls_flags", "independent_segments", "-hls_segment_type", "mpegts",
				"-hls_segment_filename", fmt.Sprintf("%s/playlist_%%v/data%%02d.ts", outputDir),
				"-master_pl_name", "playlist.m3u8",
				"-var_stream_map", "v:0 v:1 v:2 v:3",
				fmt.Sprintf("%s/playlist_%%v/manifest.m3u8", outputDir),
			)
		}
	}

	output, err := ffmpegCmd.CombinedOutput()
	log.Debugf("failed to create HLS: %v\nOutput: %s", err, string(output))
	return err
}
