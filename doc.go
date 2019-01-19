
//此源码被清华学神尹成大魔王专业翻译分析并修改
//尹成QQ77025077
//尹成微信18510341407
//尹成所在QQ群721929980
//尹成邮箱 yinc13@mails.tsinghua.edu.cn
//尹成毕业于清华大学,微软区块链领域全球最有价值专家
//https://mvp.microsoft.com/zh-cn/PublicProfile/4033620
/*
IPFS是一个全局的、版本控制的对等文件系统

IPFS包中有用于各种低级的子包
公用设施，依次组装成：

核心/…
  为消费者提供所需的所有旋钮的低级API，
  我们努力保持稳定。
壳牌/…
  高级API，让用户轻松访问公共
  操作（例如，从读卡器创建文件节点而不进行包装
  使用元数据）。我们努力保持稳定。

然后在核心之上。和壳牌/…Go API，我们有：

CMD/…
  命令行可执行文件
测试/…
  集成测试等。

为了避免周期性进口，进口不应拉高进口水平。
将API放入较低级别的包中。例如，您可以导入
命令中的核心和外壳。或测试/…，但无法导入任何
来自核心的外壳。
**/

package ipfs