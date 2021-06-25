# go-snark 
zkSNARK library implementation in Go



### CLI usage
*The cli still needs some improvements, such as seting input files, etc.*

#In this example we will follow the equation example from [Vitalik](https://medium.com/@VitalikButerin/quadratic-arithmetic-programs-from-zero-to-hero-f6d558cea649)'s article: `y = x^3 + x + 5`, where `y==35` and `x==3`. So we want to prove that we know a secret `x` such as the result of the equation is `35`.
# 证明存在一个x使得 等式`y = x^3 + x + 5`中的 y==35，也就是公开y=35，不公开x=3，证明等式成立

#### Compile circuit
#下述三个文件需要提前生成，并将路径复制到需要的区域
Having a circuit file `test.tx`:
```
func exp3(private a):
	b = a * a
	c = a * b
	return c

func main(private s0, public s1):
	s3 = exp3(s0)
	s4 = s3 + s0
	s5 = s4 + 5
	equals(s1, s5)
	out = 1 * 1
```
And a private inputs file `privateInputs.json`
```
[
	3
]
```
And a public inputs file `publicInputs.json`
```
[
	35
]
```

In the command line, execute:
#这条命令不能生成`compiledcircuit.json`文件，后续命令会找不到这个文件导致程序不能运行,
# main.exe是go build main.go生成的可执行文件
```
> main.exe compile D:\workspace\go-snark-master\cli\test.tx
```
If you want to have the wasm input ready also, add the flag `wasm`
#这条命令可以生成`compiledcircuit.json`文件
```
> main.exe compile D:\workspace\go-snark-master\cli\test.tx wasm
```

This will output the `compiledcircuit.json` file.

#### Trusted Setup
Having the `compiledcircuit.json`, now we can generate the `TrustedSetup`:
```
> main.exe trustedsetup
```
This will create the file `trustedsetup.json` with the TrustedSetup data, and also a `toxic.json` file, with the parameters to delete from the `Trusted Setup`.

If you want to have the wasm input ready also, add the flag `wasm`
```
> main.exe trustedsetup wasm
```

#### Generate Proofs
Assumming that we have the `compiledcircuit.json`, `trustedsetup.json`, `privateInputs.json` and the `publicInputs.json` we can now generate the `Proofs` with the following command:
```
> main.exe genproofs
```

This will store the file `proofs.json`, that contains all the SNARK proofs.

#### Verify Proofs
Having the `proofs.json`, `compiledcircuit.json`, `trustedsetup.json` `publicInputs.json` files, we can now verify the `Pairings` of the proofs, in order to verify the proofs.
```
> main.exe verify
```
This will return a `true` if the proofs are verified, or a `false` if the proofs are not verified.

### Cli using Groth16
All this process can be done using protocol:
# groth16算法的零知识证明使用如下：
```
> main.exe compile test.tx
> main.exe groth16 trustedsetup
> main.exe groth16 genproofs
> main.exe groth16 verify
```



# 代码上传命令
# 
（1）（ git add .），若遇到如图3中的问题，是因为没有github库的地址，可以采用（git remote add origin https://github.com/hudiankai/Golearn.git）
的命令重新设定相关github库的地址，后面的https是相关的地址。（注释：网上有说在git clone + https地址时可能会超时，相关办法是把https变为git）
（2）（git commit -m "相关备注"），
（3）（git pull），拉取github上的最新代码，遇到图3中的超时问题，
解决方法：如下图所示，与（git push）超时一样，采用命令GOPROXY设定代理
（set GOPROXY=https://goproxy.io,https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct）
（4）（git push）会弹出相关git界面，需要输入账号和密码，