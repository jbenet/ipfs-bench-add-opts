package main

import (
  "strings"
  "os/exec"
  "fmt"
  "time"
  "os"
)

var datastores = map[string]string{
  "flatfs": "",
  "badger": "--profile=badgerds",
}

var layouts = map[string]string{
  "normal":  "",
  "trickle": "--trickle",
}

var chunkers = map[string]string{
  "size":   "size-262144",
  "rabin1": "rabin-512-1024-2048",
  "rabin2": "rabin-512-1024-65536",
}

func die(err error) {
  fmt.Fprintf(os.Stderr, "error: %v", err)
  os.Exit(1)
}

func cmd(c string, env string) error {
  c = strings.Replace(c, "  ", " ", -1)
  cs := strings.Split(c, " ")
  cmd := exec.Command(cs[0], cs[1:]...)
  if env != "" {
    cmd.Env = append(cmd.Env, env)
  }

  out, err := cmd.CombinedOutput()

  fmt.Println("```")
  fmt.Println(">", env, c)
  fmt.Print(string(out))
  if out[len(out) - 1] != '\n' {
    fmt.Println("")
  }
  fmt.Println("```")

  return err
}

type test struct {
  RepoPath  string
  Datastore string
  Layout    string
  Chunker   string
  FilePath  string

  TAdd time.Duration
}

func (t *test) cmd(c string) error {
  return cmd(c, "IPFS_PATH=" + t.RepoPath)
}

func (t *test) repoInit() error {
  c := "ipfs init"
  if t.Datastore != "flatfs" {
    c += " " + datastores[t.Datastore]
  }

  return t.cmd(c)
}

func (t *test) add(path string) error {
  tstart := time.Now()

  add := "ipfs add -Q --chunker=%s -r %s %s"
  c := fmt.Sprintf(add, chunkers[t.Chunker], layouts[t.Layout], path)
  err := t.cmd(c)

  tend := time.Now()
  t.TAdd = tend.Sub(tstart)
  return err
}

func (t *test) stats() error {
  fmt.Println("add took:", t.TAdd)
  t.cmd("du -sh " + t.RepoPath)
  t.cmd("ipfs repo stat")
  return nil
}

func (t *test) run() error {

  fmt.Printf("### {%s, %s, %s}\n", t.Datastore, t.Chunker, t.Layout)

  fmt.Println("Options:")
  fmt.Println("- Datastore:", t.Datastore)
  fmt.Println("- Chunker:", t.Chunker)
  fmt.Println("- Layout:", t.Layout)

  if err := t.repoInit(); err != nil {
    return err
  }

  if err := t.add(t.FilePath); err != nil {
    return err
  }

  if err := t.stats(); err != nil {
    return err
  }

  return nil
}

func run(path string) error {

  fmt.Println("---")
  fmt.Println("##", time.Now().Format("2006-01-02 15:04:05"))
  fmt.Printf("benchmarking ipfs with directory: `%s`\n", path)

  if err := cmd("du -sh " + path, ""); err != nil {
    return err
  }

  // if err := cmd(fmt.Sprintf("bash -c 'find %s | wc'", path), ""); err != nil {
  //   return err
  // }

  for dk := range datastores {
    for ck := range chunkers {
      for lk := range layouts {

        t := test{
          RepoPath:  fmt.Sprintf("ipfs-repo-%s.%s.%s", dk, ck, lk),
          Datastore: dk,
          Layout:    lk,
          Chunker:   ck,
          FilePath:  path,
        }

        err := t.run()
        if err != nil {
          return err
        }
      }
    }
  }

  return nil
}

func main() {
  if len(os.Args) < 2 {
    fmt.Printf("%s <path-to-test-files>\n", os.Args[0])
    fmt.Println("benchmark ipfs with directory")
    return
  }

  err := run(os.Args[1])
  if err != nil {
    die(err)
  }
}
