# xdxct-vgpu-device-manager

## Description
The vGPU Device Manager is a tool designed for system administrators to make working with vGPU devices easier.
It allows administrators to declaratively define a set of possible vGPU device configurations they would like applied to all GPUs on a node.

## Build xgv-vgpu-dm
This will generate a binary called xgv-vgpu-dm.
```shell
make cmds-nvidia-vgpu-dm
```

## Usage
1. Create a specific vGPU device from a configuration file.
```shell
sudo ./xgv-vgpu-dm apply -f examples/config-vgpu.yaml -c PANGU-A0-1G-1-CORE
```
2. Create a specific vGPU device with debug out.
```shell
sudo ./xgv-vgpu-dm -v apply -f examples/config-vgpu.yaml -c PANGU-A0-128M-1-CORE
```

## Kubernetes Deployment
