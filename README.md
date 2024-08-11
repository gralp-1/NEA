

# dev steps
- [x] Early switch from imgui to raylib for flexibility reasons
- [x] factor out the texture send from applying the filters for **State diffing**
- [x] fix issue related to when you copy the image from the original to the shown
- [ ] port backend to go image api, convert to raylib images before sending to GPU
- [ ] implement state filter diffing so filters only need to be reapplied if the filters change, not every frame
  - Explain performance implications, include benchmarks of applying all filters (worst case) and memory stress of copying the image every fram
  - 77ms (13fps) to do it every frame
- [ ] implement a way to save the image to disk
- [ ] widescale logging for debugging
- [ ] Testing "raw" vs native way of sending data to GPU
  - Revisit after backend change
- [ ] Benchmarking to assess 60FPS target
- [ ] Custom dynamic GUI builder based on giu library


```go
// raw way
length := state.ShownImage.Width * state.ShownImage.Height
slice := (*[1 << 30]color.RGBA)(unsafe.Pointer(state.ShownImage))[:length:length]
rl.UpdateTexture(state.CurrentTexture, slice)
```

```go
// native way (could be slower)
state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)
```
