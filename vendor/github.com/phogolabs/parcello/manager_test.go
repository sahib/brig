package parcello_test

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	zipexe "github.com/daaku/go.zipexe"
	"github.com/kardianos/osext"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/parcello/fake"
)

var _ = Describe("ResourceManager", func() {
	var (
		manager  *parcello.ResourceManager
		resource *parcello.Resource
		bundle   *parcello.Bundle
	)

	BeforeEach(func() {
		var err error

		compressor := parcello.ZipCompressor{
			Config: &parcello.CompressorConfig{
				Logger:   ioutil.Discard,
				Filename: "bundle",
				Recurive: true,
			},
		}

		fileSystem := parcello.Dir("./fixture")

		ctx := &parcello.CompressorContext{
			FileSystem: fileSystem,
		}

		bundle, err = compressor.Compress(ctx)
		Expect(err).NotTo(HaveOccurred())

		manager = &parcello.ResourceManager{}
	})

	JustBeforeEach(func() {
		resource = parcello.BinaryResource(bundle.Body)
		Expect(manager.Add(resource)).To(Succeed())
	})

	Describe("NewResourceManager", func() {
		var (
			name       string
			fileSystem parcello.FileSystem
		)

		BeforeEach(func() {
			path, err := osext.Executable()
			Expect(err).To(Succeed())

			path, name = filepath.Split(path)
			fileSystem = parcello.Dir(path)
		})

		It("creates new manager successfully", func() {
			cfg := &parcello.ResourceManagerConfig{
				Path:       name,
				FileSystem: fileSystem,
			}
			m, err := parcello.NewResourceManager(cfg)
			Expect(m).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when file stat fails", func() {
			BeforeEach(func() {
				exec := &fake.File{}
				exec.StatReturns(nil, fmt.Errorf("oh no!"))

				fs := &fake.FileSystem{}
				fs.OpenFileReturns(exec, nil)

				fileSystem = fs
			})

			It("returns the error", func() {
				cfg := &parcello.ResourceManagerConfig{
					Path:       name,
					FileSystem: fileSystem,
				}

				m, err := parcello.NewResourceManager(cfg)
				Expect(m).To(BeNil())
				Expect(err).To(MatchError("oh no!"))
			})
		})
	})

	Describe("Add", func() {
		Context("when the resource is added second time", func() {
			It("returns an error", func() {
				Expect(manager.Add(resource)).To(MatchError("invalid path: 'resource/reports/2018.txt'"))
			})
		})

		Context("when the algorithm is unsupported", func() {
			JustBeforeEach(func() {
				manager = &parcello.ResourceManager{}
				manager.NewReader = func(r io.ReaderAt, s int64) (*zip.Reader, error) {
					reader, err := zipexe.NewReader(r, s)
					if err != nil {
						return nil, err
					}
					reader.File[0].FileHeader.Method = 2000
					return reader, nil
				}
			})

			It("returns an error", func() {
				Expect(manager.Add(resource)).To(MatchError("zip: unsupported compression algorithm"))
			})
		})

		Context("when the file is corrupted", func() {
			JustBeforeEach(func() {
				manager = &parcello.ResourceManager{}
				manager.NewReader = func(r io.ReaderAt, s int64) (*zip.Reader, error) {
					reader, err := zipexe.NewReader(r, s)
					if err != nil {
						return nil, err
					}
					reader.File[0].FileHeader.CRC32 = 123
					return reader, nil
				}
			})

			It("returns an error", func() {
				Expect(manager.Add(resource)).To(MatchError("zip: checksum error"))
			})
		})

		Context("when the resource is not zip", func() {
			It("returns an error", func() {
				Expect(manager.Add(parcello.BinaryResource([]byte("lol")))).To(MatchError("Couldn't Open As Executable"))
			})

			It("panics", func() {
				Expect(func() { parcello.AddResource([]byte("lol")) }).To(Panic())
			})
		})
	})

	Describe("Dir", func() {
		It("returns a valid sub-manager", func() {
			group, err := manager.Dir("/resource")
			Expect(err).To(BeNil())

			file, err := group.Open("/reports/2018.txt")
			Expect(file).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())

			data, err := ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("Report 2018\n"))
		})

		Context("when group is a file not a directory", func() {
			It("returns an error", func() {
				group, err := manager.Dir("/resource/reports/2018.txt")
				Expect(group).To(BeNil())
				Expect(err).To(MatchError(os.ErrNotExist))
			})
		})

		Context("when the manager is global", func() {
			var (
				original parcello.FileSystemManager
				manager  *fake.FileSystemManager
			)

			BeforeEach(func() {
				manager = &fake.FileSystemManager{}

				original = parcello.Manager
				parcello.Manager = manager
			})

			AfterEach(func() {
				parcello.Manager = original
			})

			It("returns a sub-manager", func() {
				manager.DirReturns(manager, nil)
				Expect(parcello.ManagerAt("/nil")).To(Equal(parcello.Manager))
			})

			Context("when the directory does not exist", func() {
				It("panics", func() {
					manager.DirReturns(nil, fmt.Errorf("oh no!"))
					Expect(func() { parcello.ManagerAt("/i/do/not/exist") }).To(Panic())
				})
			})
		})
	})

	Describe("Open", func() {
		It("opens the root successfully", func() {
			file, err := manager.Open("/")
			Expect(file).NotTo(BeNil())
			Expect(err).To(BeNil())
		})

		Context("when the resource is empty", func() {
			It("returns an error", func() {
				file, err := manager.Open("/migration.sql")
				Expect(file).To(BeNil())
				Expect(err).To(MatchError("open /migration.sql: file does not exist"))
			})
		})

		Context("when the file is directory", func() {
			It("returns an error", func() {
				file, err := manager.Open("/resource/reports")
				Expect(file).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		Context("when the global resource is empty", func() {
			It("returns an error", func() {
				file, err := parcello.Open("migration.sql")
				Expect(file).To(BeNil())
				Expect(err).To(MatchError("open migration.sql: file does not exist"))
			})
		})

		It("returns the resource successfully", func() {
			file, err := manager.Open("/resource/reports/2018.txt")
			Expect(file).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())

			data, err := ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("Report 2018\n"))
		})

		Context("when the file is open more than once for read", func() {
			It("does not change the mod time", func() {
				file, err := manager.Open("/resource/reports/2018.txt")
				Expect(file).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				info, err := file.Stat()
				Expect(err).NotTo(HaveOccurred())

				file, err = manager.Open("/resource/reports/2018.txt")
				Expect(file).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				info2, err := file.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(info.ModTime()).To(Equal(info2.ModTime()))
			})
		})

		It("returns a readonly resource", func() {
			file, err := manager.Open("/resource/reports/2018.txt")
			Expect(file).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())

			_, err = fmt.Fprintln(file.(io.Writer), "hello")
			Expect(err).To(MatchError("File is read-only"))
		})

		Context("when the file with the requested name does not exist", func() {
			It("returns an error", func() {
				file, err := manager.Open("/resource/migration.sql")
				Expect(file).To(BeNil())
				Expect(err).To(MatchError("open /resource/migration.sql: file does not exist"))
			})
		})
	})

	Describe("OpenFile", func() {
		Context("when the file does not exist", func() {
			It("creates the file", func() {
				file, err := manager.OpenFile("/resource/secrets.txt", os.O_CREATE, 0600)
				Expect(file).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the file is directory", func() {
			It("returns an error", func() {
				file, err := manager.OpenFile("/resource/reports", os.O_RDWR, 0600)
				Expect(file).To(BeNil())
				Expect(err).To(MatchError("open /resource/reports: Is directory"))
			})
		})

		Context("when the file exists", func() {
			It("truncs the file content", func() {
				file, err := manager.OpenFile("/resource/reports/2018.txt", os.O_CREATE|os.O_TRUNC, 0600)
				Expect(file).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(BeEmpty())
			})

			Context("when the file is open more than once for write", func() {
				It("does not change the mod time", func() {
					start := time.Now()

					file, err := manager.OpenFile("/resource/reports/2018.txt", os.O_WRONLY, 0600)
					Expect(file).NotTo(BeNil())
					Expect(err).NotTo(HaveOccurred())

					info, err := file.Stat()
					Expect(err).NotTo(HaveOccurred())
					modTime := info.ModTime()

					Expect(modTime.After(start)).To(BeTrue())
				})
			})

			Context("when the os.O_TRUNC flag is not provided", func() {
				It("returns an error", func() {
					file, err := manager.OpenFile("/resource/reports/2018.txt", os.O_CREATE, 0600)
					Expect(file).To(BeNil())
					Expect(err).To(MatchError("open /resource/reports/2018.txt: file already exists"))
				})
			})

			Context("when the file is open for append", func() {
				It("appends content successfully", func() {
					file, err := manager.OpenFile("/resource/reports/2018.txt", os.O_RDWR|os.O_APPEND, 0600)
					Expect(file).NotTo(BeNil())
					Expect(err).NotTo(HaveOccurred())

					_, err = fmt.Fprint(file, "hello")
					Expect(err).NotTo(HaveOccurred())

					_, err = file.Seek(0, io.SeekStart)
					Expect(err).NotTo(HaveOccurred())

					data, err := ioutil.ReadAll(file)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(data)).To(Equal("Report 2018\nhello"))
				})
			})

			Context("when the file is open for WRITE only", func() {
				Context("when we try to read", func() {
					It("returns an error", func() {
						file, err := manager.OpenFile("/resource/reports/2018.txt", os.O_WRONLY, 0600)
						Expect(file).NotTo(BeNil())
						Expect(err).NotTo(HaveOccurred())

						_, err = ioutil.ReadAll(file)
						Expect(err).To(MatchError("File is write-only"))
					})
				})
			})
		})
	})

	Describe("Walk", func() {
		Context("when the resource is empty", func() {
			It("returns an error", func() {
				err := manager.Walk("/documents", func(path string, info os.FileInfo, err error) error {
					return nil
				})

				Expect(err).To(MatchError(os.ErrNotExist))
			})
		})

		Context("when the resource has hierarchy of directories and files", func() {
			It("walks through all of them", func() {
				paths := []string{}
				err := manager.Walk("/", func(path string, info os.FileInfo, err error) error {
					paths = append(paths, path)
					return nil
				})

				Expect(paths).To(HaveLen(11))
				Expect(paths[0]).To(Equal("/"))
				Expect(paths[1]).To(Equal("/resource"))
				Expect(paths[2]).To(Equal("/resource/reports"))
				Expect(paths[3]).To(Equal("/resource/reports/2018.txt"))
				Expect(paths[4]).To(Equal("/resource/scripts"))
				Expect(paths[5]).To(Equal("/resource/scripts/schema.sql"))
				Expect(paths[6]).To(Equal("/resource/templates"))
				Expect(paths[7]).To(Equal("/resource/templates/html"))
				Expect(paths[8]).To(Equal("/resource/templates/html/index.html"))
				Expect(paths[9]).To(Equal("/resource/templates/yml"))
				Expect(paths[10]).To(Equal("/resource/templates/yml/schema.yml"))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the start node is file", func() {
				It("walks through the file only", func() {
					cnt := 0
					err := manager.Walk("/resource/reports/2018.txt", func(path string, info os.FileInfo, err error) error {
						cnt = cnt + 1
						Expect(path).To(Equal("/resource/reports/2018.txt"))
						Expect(info.Name()).To(Equal("2018.txt"))
						Expect(info.Size()).NotTo(BeZero())
						return nil
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(cnt).To(Equal(1))
				})
			})

			It("walks through all of root children", func() {
				cnt := 0
				paths := []string{}
				err := manager.Walk("/resource/templates", func(path string, info os.FileInfo, err error) error {
					paths = append(paths, path)
					cnt = cnt + 1
					return nil
				})

				Expect(paths).To(HaveLen(5))
				Expect(paths[0]).To(Equal("/resource/templates"))
				Expect(paths[1]).To(Equal("/resource/templates/html"))
				Expect(paths[2]).To(Equal("/resource/templates/html/index.html"))
				Expect(paths[3]).To(Equal("/resource/templates/yml"))
				Expect(paths[4]).To(Equal("/resource/templates/yml/schema.yml"))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when the walker returns an error", func() {
				It("returns the error", func() {
					err := manager.Walk("/resource", func(path string, info os.FileInfo, err error) error {
						return fmt.Errorf("Oh no!")
					})

					Expect(err).To(MatchError("Oh no!"))
				})

				Context("when the walk returns an error for sub-directory", func() {
					It("returns the error", func() {
						err := manager.Walk("/resource", func(path string, info os.FileInfo, err error) error {
							if path == "/resource/templates" {
								return fmt.Errorf("Oh no!")
							}
							return nil
						})

						Expect(err).To(MatchError("Oh no!"))
					})
				})
			})
		})
	})
})

var _ = Describe("DefaultManager", func() {
	It("creates a new manager successfully", func() {
		manager := parcello.DefaultManager(osext.Executable)
		Expect(manager).NotTo(BeNil())
		_, ok := manager.(*parcello.ResourceManager)
		Expect(ok).To(BeTrue())
	})

	Context("when the executable cannot be found", func() {
		It("panics", func() {
			fn := func() (string, error) { return "", fmt.Errorf("oh no!") }
			Expect(func() { parcello.DefaultManager(fn) }).To(Panic())
		})
	})

	Context("when the filesystem fails", func() {
		It("panics", func() {
			fn := func() (string, error) { return "/i/do/not/exist", nil }
			Expect(func() { parcello.DefaultManager(fn) }).To(Panic())
		})
	})

	Context("when dev mode is enabled", func() {
		BeforeEach(func() {
			os.Setenv("PARCELLO_DEV_ENABLED", "1")
		})

		AfterEach(func() {
			os.Unsetenv("PARCELLO_DEV_ENABLED")
		})

		It("creates a new dir manager", func() {
			manager := parcello.DefaultManager(osext.Executable)
			Expect(manager).NotTo(BeNil())
			dir, ok := manager.(parcello.Dir)
			Expect(ok).To(BeTrue())
			Expect(string(dir)).To(Equal("."))
		})

		Context("when the directory is provided", func() {
			BeforeEach(func() {
				os.Setenv("PARCELLO_RESOURCE_DIR", "./root")
			})

			AfterEach(func() {
				os.Unsetenv("PRACELLO_RESOURCE_DIR")
			})

			It("creates a new dir manager", func() {
				manager := parcello.DefaultManager(osext.Executable)
				Expect(manager).NotTo(BeNil())
				dir, ok := manager.(parcello.Dir)
				Expect(ok).To(BeTrue())
				Expect(string(dir)).To(Equal("./root"))
			})
		})
	})
})
