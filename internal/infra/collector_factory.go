package infra

type CollectorFactory struct{}

func (f *CollectorFactory) CreateCollectors() []interface{} {
    var collectors []interface{}

    collectors = append(collectors, NewCPUMemoryCollector())

    if hasNvidiaGPU() {
        collectors = append(collectors, NewNvidiaGPUCollector())
    }
    // TODO: Implement AMD GPU detection and collection in the future
    // else if hasAMDGPU() {
    //     collectors = append(collectors, NewAMDGPUCollector())
    // }

    return collectors
}

func hasNvidiaGPU() bool {
    // Implementation to detect NVIDIA GPU
    return false
}

// TODO: Implement AMD GPU detection in the future
// func hasAMDGPU() bool {
//     // Implementation to detect AMD GPU
//     return false
// }
