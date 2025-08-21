
# 安装库
# pip install bio-isoelectric-point
from isoelectric import ipc
import sys

input = sys.argv[1]

try:
    # 使用isoelectric库的predict_isoelectric_point函数
    # 采用'IPC_protein'模型计算等电点
    # 返回数值或出错返回0
    result = ipc.predict_isoelectric_point(input, 'IPC_protein')
    print(result)
except:
    print(0)
