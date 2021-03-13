# LinuxMem

> Easilie read and write to the memory of process on a linux system. 

## Example

```go
// PID
mem := linuxmem.New(4036, false)

// Buffer
bs := make([]byte, 4)
binary.LittleEndian.PutUint32(bs, 99)

// Address
mem.WriteMemory(0x7ffc9c09a294, bs)
```
