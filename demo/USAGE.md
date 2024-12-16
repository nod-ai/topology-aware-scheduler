# Topology-Aware GPU Cluster Scheduler Demo

A simulation and visualization tool demonstrating topology-aware GPU scheduling in a leaf-spine network architecture.

## Features
- Real-time cluster state visualization
- Dynamic job scheduling and resource allocation
- Topology-aware placement optimization
- Interactive simulation controls
- Performance metrics tracking

## Requirements
```
streamlit
plotly
numpy
pandas
```

## Installation
```bash
python3 -m pip install streamlit plotly numpy pandas
```

## Usage
```bash
python3 -m streamlit run demo.py
```

## Architecture
The scheduler manages a 128-node cluster with:
- 4 GPUs per node (512 GPUs total)
- 32 leaf switches (4 nodes each)
- Central spine switch

## Dashboard Components
- Cluster Topology Heatmap
- GPU Utilization Graph
- Job Queue Monitoring
- Resource Usage Metrics

## Simulation Controls
- Adjust job arrival rate
- Configure job sizes
- Control update interval

## License
MIT License

## Contributing
Pull requests welcome. For major changes, open an issue first.
