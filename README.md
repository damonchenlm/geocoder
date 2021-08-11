## 坐标解析
### 1.访问 show-after-geocode.html
### 2.选择文件 -> 上传格式与input.csv相符的文件 -> geocode
### 3.解析完成后，点击 output 输出结果文件
### 4.解析完成后，点击 show in the map 在地图上显示坐标点

## 直接投图
### 1.访问 show-with-result-file.html
### 2.选择文件 -> 上传格式与res.csv相符的文件 -> show in the map

## Tips
### 1.确保输入地点没有特殊符号(用以构成url访问链接)
### 2.结果为 valid 不能保证为正确结果，可能是模糊地点(多个类似结果中的一个)
### 3.确保输入文件格式，或自行修改 getFile()中 Papa.Parse()内容
### 4.若地图显示不完全，可通过改变窗口大小或用F12呼出开发工具来触发重新渲染