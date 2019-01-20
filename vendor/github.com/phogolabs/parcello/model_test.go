package parcello_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/parcello"
)

var _ = Describe("Model", func() {
	Describe("ResourceFileInfo", func() {
		var (
			info *parcello.ResourceFileInfo
			node *parcello.Node
		)

		BeforeEach(func() {
			data := []byte("hello")

			node = &parcello.Node{
				Name:    "node",
				ModTime: time.Now(),
				Mutex:   &sync.RWMutex{},
				IsDir:   false,
				Content: &data,
			}

			info = &parcello.ResourceFileInfo{Node: node}
		})

		It("returns the Name successfully", func() {
			Expect(info.Name()).To(Equal("node"))
		})

		It("returns the Size successfully", func() {
			Expect(info.Size()).To(Equal(int64(len(*node.Content))))
		})

		It("returns the Mode successfully", func() {
			Expect(info.Mode()).To(BeZero())
		})

		It("returns the ModTime successfully", func() {
			Expect(info.ModTime()).To(Equal(node.ModTime))
		})

		It("returns the IsDir successfully", func() {
			Expect(info.IsDir()).To(BeFalse())
		})

		It("returns the Sys successfully", func() {
			Expect(info.Sys()).To(BeNil())
		})
	})

	Describe("ResourceFile", func() {
		var (
			file *parcello.ResourceFile
			node *parcello.Node
		)

		Context("when the node is file", func() {
			BeforeEach(func() {
				data := []byte("hello")

				node = &parcello.Node{
					Name:    "sample.txt",
					ModTime: time.Now(),
					Mutex:   &sync.RWMutex{},
					IsDir:   false,
					Content: &data,
				}

				file = parcello.NewResourceFile(node)

				_, err := file.Seek(int64(len(data)), io.SeekStart)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads successfully", func() {
				_, err := file.Seek(0, io.SeekStart)
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("hello"))
			})

			It("writes successfully", func() {
				fmt.Fprintf(file, ",jack")

				_, err := file.Seek(0, io.SeekStart)
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("hello,jack"))
			})

			It("closes successfully", func() {
				Expect(file.Close()).To(Succeed())
			})

			It("seeks successfully", func() {
				n, err := file.Seek(1, 0)
				Expect(err).To(BeNil())
				Expect(n).To(Equal(int64(1)))
			})

			It("reads the directory fails", func() {
				files, err := file.Readdir(-1)
				Expect(err).To(MatchError("Not supported"))
				Expect(files).To(HaveLen(0))
			})

			It("returns the information successfully", func() {
				info, err := file.Stat()
				Expect(err).To(BeNil())
				Expect(info.IsDir()).To(BeFalse())
				Expect(info.Name()).To(Equal("sample.txt"))
			})
		})

		Context("when the node is directory", func() {
			BeforeEach(func() {
				data1 := []byte("hello")
				data2 := []byte("world")
				node = &parcello.Node{
					Name:  "documents",
					IsDir: true,
					Children: []*parcello.Node{
						{
							Name:    "sample.txt",
							Content: &data1,
						},
						{
							Name:    "report.txt",
							Content: &data2,
						},
					},
				}

				file = parcello.NewResourceFile(node)
			})

			It("reads the directory successfully", func() {
				files, err := file.Readdir(-1)
				Expect(err).To(BeNil())
				Expect(files).To(HaveLen(2))

				info := files[0]
				Expect(info.Name()).To(Equal("sample.txt"))

				info = files[1]
				Expect(info.Name()).To(Equal("report.txt"))
			})

			Context("when the n is 1", func() {
				It("reads the directory successfully", func() {
					files, err := file.Readdir(1)
					Expect(err).To(BeNil())
					Expect(files).To(HaveLen(1))

					info := files[0]
					Expect(info.Name()).To(Equal("sample.txt"))
				})
			})

			It("returns the information successfully", func() {
				info, err := file.Stat()
				Expect(err).To(BeNil())
				Expect(info.IsDir()).To(BeTrue())
				Expect(info.Name()).To(Equal("documents"))
				Expect(info.Size()).To(BeZero())
			})
		})
	})
})
