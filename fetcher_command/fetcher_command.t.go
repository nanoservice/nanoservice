// Code generated by ./_fetcher.tt.go
package fetcher_command

import ("github.com/nanoservice/nanoservice/command")

type Fetcher struct {
  container *map[string]command.Command
  defaultValue *command.Command
}

func New(container *map[string]command.Command) *Fetcher {
  return &Fetcher{
    container:    container,
    defaultValue: nil,
  }
}

func (f *Fetcher) WithDefault(value *command.Command) *Fetcher {
  f.defaultValue = value
  return f
}

func (f *Fetcher) Fetch(key string) command.Command {
  if f.defaultValue == nil {
    return f.HardFetch(key)
  }
  return f.FetchWithDefault(key, *f.defaultValue)
}

func (f *Fetcher) HardFetch(key string) command.Command {
  return (*f.container)[key]
}

func (f *Fetcher) FetchWithDefault(key string, defaultValue command.Command) command.Command {
  if result, found := (*f.container)[key]; found {
    return result
  }
  return defaultValue
}
