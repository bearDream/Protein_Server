# 导入下下来的plot库 
from plot import plot
import sys

path = sys.argv[1]
# 使用plot库分析PDB文件中的二面角
# 生成带色谱的散点图，显示氨基酸构象分布
# 输出为JPG格式
plot("./static/models/"+path+".pdb", cmap='viridis', alpha=0.75,
     dpi=100, save=True, show=False, out="./static/ramachandran_plots/"+path+".jpg")
