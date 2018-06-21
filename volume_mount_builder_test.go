package piper_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/joshzarrabi/piper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VolumeMountBuilder", func() {
	var (
		builder         piper.VolumeMountBuilder
		localMountPoint string
		mounts          []piper.DockerVolumeMount
		err             error
		input1          string
		input2          string
	)

	Describe("Build", func() {
		BeforeEach(func() {
			input1, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			input2, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			os.RemoveAll(localMountPoint)
			os.RemoveAll(input1)
			os.RemoveAll(input2)
		})

		It("copies the inputs to a temporary directory", func() {
			_, localMountPoint, err = builder.Build([]piper.VolumeMount{
				piper.VolumeMount{Name: "input-1"},
				piper.VolumeMount{Name: "input-2"},
				piper.VolumeMount{Name: "output-1"},
				piper.VolumeMount{Name: "output-2"},
			}, []string{
				fmt.Sprintf("input-1=%s", input1),
				fmt.Sprintf("input-2=%s", input2),
			}, []string{
				"output-1=/some/path-3",
				"output-2=/some/path-4",
			})
			Expect(err).NotTo(HaveOccurred())
			inputDirs, err := ioutil.ReadDir(localMountPoint)
			Expect(err).NotTo(HaveOccurred())
			dirNames := []string{}
			for _, dir := range inputDirs {
				dirNames = append(dirNames, dir.Name())
				Expect(dir.IsDir()).To(BeTrue())
			}
			Expect(dirNames).To(ConsistOf(filepath.Base(input1), filepath.Base(input2)))
		})

		It("builds the volume mounts", func() {
			mounts, localMountPoint, err = builder.Build([]piper.VolumeMount{
				piper.VolumeMount{Name: "input-1"},
				piper.VolumeMount{Name: "input-2"},
				piper.VolumeMount{Name: "output-1"},
				piper.VolumeMount{Name: "output-2"},
			}, []string{
				fmt.Sprintf("input-1=%s", input1),
				fmt.Sprintf("input-2=%s", input2),
			}, []string{
				"output-1=/some/path-3",
				"output-2=/some/path-4",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(mounts).To(Equal([]piper.DockerVolumeMount{
				{
					LocalPath:  filepath.Join(localMountPoint, filepath.Base(input1)),
					RemotePath: "/tmp/build/input-1",
				},
				{
					LocalPath:  filepath.Join(localMountPoint, filepath.Base(input2)),
					RemotePath: "/tmp/build/input-2",
				},
				{
					LocalPath:  "/some/path-3",
					RemotePath: "/tmp/build/output-1",
				},
				{
					LocalPath:  "/some/path-4",
					RemotePath: "/tmp/build/output-2",
				},
			}))
		})

		It("honors the path given in the VolumeMount", func() {
			mounts, localMountPoint, err = builder.Build([]piper.VolumeMount{
				piper.VolumeMount{Name: "input-1", Path: "some/path/to/input"},
				piper.VolumeMount{Name: "input-2"},
				piper.VolumeMount{Name: "output-1"},
				piper.VolumeMount{Name: "output-2", Path: "some/path/to/output"},
			}, []string{
				fmt.Sprintf("input-1=%s", input1),
				fmt.Sprintf("input-2=%s", input2),
			}, []string{
				"output-1=/some/path-3",
				"output-2=/some/path-4",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(mounts).To(Equal([]piper.DockerVolumeMount{
				{
					LocalPath:  filepath.Join(localMountPoint, filepath.Base(input1)),
					RemotePath: "/tmp/build/some/path/to/input",
				},
				{
					LocalPath:  filepath.Join(localMountPoint, filepath.Base(input2)),
					RemotePath: "/tmp/build/input-2",
				},
				{
					LocalPath:  "/some/path-3",
					RemotePath: "/tmp/build/output-1",
				},
				{
					LocalPath:  "/some/path-4",
					RemotePath: "/tmp/build/some/path/to/output",
				},
			}))
		})

		Context("failure cases", func() {
			Context("when the input pairs are malformed", func() {
				It("returns an error", func() {
					_, _, err := builder.Build([]piper.VolumeMount{}, []string{
						"input-1=something",
						"input-2",
					}, []string{})
					Expect(err).To(MatchError("could not parse input \"input-2\". must be of form <input-name>=<input-location>"))
				})
			})

			Context("when an input pair is not specified, but is required", func() {
				It("returns an error", func() {
					_, _, err := builder.Build([]piper.VolumeMount{
						{Name: "input-1"},
						{Name: "input-2"},
						{Name: "input-3"},
					}, []string{
						fmt.Sprintf("input-1=%s", input1),
					}, []string{})
					Expect(err).To(MatchError(`The following required inputs/outputs are not satisfied: input-2, input-3.`))
				})
			})

			Context("when the output pairs are malformed", func() {
				It("returns an error", func() {
					_, _, err := builder.Build([]piper.VolumeMount{}, []string{}, []string{
						"output-1=something",
						"output-2",
					})
					Expect(err).To(MatchError("could not parse output \"output-2\". must be of form <output-name>=<output-location>"))
				})
			})

			Context("when an input pair is not specified, but is required", func() {
				It("returns an error", func() {
					_, _, err := builder.Build([]piper.VolumeMount{
						{Name: "output-1"},
						{Name: "output-2"},
						{Name: "output-3"},
					}, []string{}, []string{
						"output-1=/some/path-1",
					})
					Expect(err).To(MatchError(`The following required inputs/outputs are not satisfied: output-2, output-3.`))
				})
			})
		})
	})
})
