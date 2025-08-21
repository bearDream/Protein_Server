from plot import plot
import sys

# 将 PDB 模型转化为结构图像（使用 PyMOL 或 matplotlib）

input = sys.argv[1]
output = sys.argv[2]

plot(input, cmap='viridis', alpha=0.75, dpi=100, save=True, show=False, out=output)