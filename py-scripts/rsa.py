# 安装库
# pip install freesasa
import freesasa
import sys

input = sys.argv[1]
path = "static/models"+input
# 使用freesasa库计算蛋白质表面积
# 计算总表面积除以残基数量
# 得到平均每个残基的溶剂可及性
try:
    structure = freesasa.Structure(path)
    result = freesasa.calc(structure)
    total_area = result.totalArea()
    r_areas = result.residueAreas()
    print(total_area / len(r_areas['A']))
except:
    print(0)


