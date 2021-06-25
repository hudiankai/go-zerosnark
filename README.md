# go-snark 
zkSNARK library implementation in Go


- `Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture`, Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, Madars Virza https://eprint.iacr.org/2013/879.pdf
- `Pinocchio: Nearly practical verifiable computation`, Bryan Parno, Craig Gentry, Jon Howell, Mariana Raykova https://eprint.iacr.org/2013/279.pdf
- `On the Size of Pairing-based Non-interactive Arguments`, Jens Groth https://eprint.iacr.org/2016/260.pdf

## Caution & Warning
Implementation of the zkSNARK  and [Groth16 protocol](https://eprint.iacr.org/2016/260.pdf) from scratch in Go done in my free time to understand the concepts. Do not use in production.

## zkSNARKs in Go
If you need to use zkSNARKs in Go, I would recommend to take a look at [go-circom-prover-verifier](https://github.com/iden3/go-circom-prover-verifier), which I've wrote using the [bn256](https://github.com/ethereum/go-ethereum/tree/master/crypto/bn256/cloudflare) for the Pairing curve operations for the Groth16 zkSNARK, and it is compatible with [circom](https://github.com/iden3/circom).

## Features
现在可以使用Pinocchio协议和Groth16协议完成完整的零知识验证。
Currently allows to do the complete path with [Pinocchio protocol](https://eprint.iacr.org/2013/279.pdf) and [Groth16 protocol](https://eprint.iacr.org/2016/260.pdf) :

0. write circuit
1. compile circuit
2. generate trusted setup
3. calculate witness
4. generate proofs
5. verify proofs

Minimal complete flow implementation（最小的完成流程）:
- [x] Finite Fields (1, 2, 6, 12) operations
- [x] G1 and G2 curve operations
- [x] BN128 Pairing
- [x] circuit flat code compiler
- [x] circuit to R1CS
- [x] polynomial operations
- [x] R1CS to QAP
- [x] generate trusted setup
- [x] generate proofs
- [x] verify proofs with BN128 pairing

## WASM usage
Experimentation with go-snark compiled to wasm: https://github.com/arnaucube/go-snark/tree/master/wasm


### CLI usage
*The cli still needs some improvements, such as seting input files, etc.*

#In this example we will follow the equation example from [Vitalik](https://medium.com/@VitalikButerin/quadratic-arithmetic-programs-from-zero-to-hero-f6d558cea649)'s article: `y = x^3 + x + 5`, where `y==35` and `x==3`. So we want to prove that we know a secret `x` such as the result of the equation is `35`.
# 证明存在一个x使得 等式`y = x^3 + x + 5`中的 y==35，也就是公开y=35，不公开x=3，证明等式成立

#### Compile circuit
#下述三个文件需要提前生成，并将路径复制到需要的区域
Having a circuit file `test.circuit`:
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
All this process can be done using [Groth16 protocol](https://eprint.iacr.org/2016/260.pdf) protocol:
# groth16算法的零知识证明使用如下：
```
> main.exe compile test.tx
> main.exe groth16 trustedsetup
> main.exe groth16 genproofs
> main.exe groth16 verify
```



### Library usage

Example:
```go
// compile circuit and get the R1CS
flatCode := `
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
`

// parse the code
parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
circuit, err := parser.Parse()
assert.Nil(t, err)
fmt.Println(circuit)


b3 := big.NewInt(int64(3))
privateInputs := []*big.Int{b3}
b35 := big.NewInt(int64(35))
publicSignals := []*big.Int{b35}

// witness
w, err := circuit.CalculateWitness(privateInputs, publicSignals)
assert.Nil(t, err)
fmt.Println("witness", w)

// now we have the witness:
// w = [1 35 3 9 27 30 35 1]

// flat code to R1CS
fmt.Println("generating R1CS from flat code")
a, b, c := circuit.GenerateR1CS()

/*
now we have the R1CS from the circuit:
a: [[0 0 1 0 0 0 0 0] [0 0 1 0 0 0 0 0] [0 0 1 0 1 0 0 0] [5 0 0 0 0 1 0 0] [0 0 0 0 0 0 1 0] [0 1 0 0 0 0 0 0] [1 0 0 0 0 0 0 0]]
b: [[0 0 1 0 0 0 0 0] [0 0 0 1 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0]]
c: [[0 0 0 1 0 0 0 0] [0 0 0 0 1 0 0 0] [0 0 0 0 0 1 0 0] [0 0 0 0 0 0 1 0] [0 1 0 0 0 0 0 0] [0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 1]]
*/


alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)


ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

// calculate trusted setup
setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)

hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)

proof, err := GenerateProofs(*circuit, setup, w, px)

b35Verif := big.NewInt(int64(35))
publicSignalsVerif := []*big.Int{b35Verif}
assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
```

##### Verify Proof generated from [snarkjs]
Is possible with `go-snark` to verify proofs generated by `snarkjs`

Example:
```go
verified, err := VerifyFromCircom("circom-test/verification_key.json", "circom-test/proof.json", "circom-test/public.json")
assert.Nil(t, err)
assert.True(t, verified)
```


## Versions
History of versions & tags of this project:
- v0.0.1: zkSnark complete flow working with Pinocchio protocol
- v0.0.2: circuit language improved (allow function calls and file imports)
- v0.0.3: Groth16 zkSnark protocol added

## Test
```
go test ./... -v
```

## vim/nvim circuit syntax highlighter
For more details and installation instructions see https://github.com/arnaucube/go-snark/tree/master/vim-syntax

---


