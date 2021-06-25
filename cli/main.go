package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	snark "github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/groth16"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/arnaucube/go-snark/utils"
	"github.com/urfave/cli"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}


func main()  {
	app := &cli.App{
		// 设定输入参数，以-Name=x的方式输入不同的参数
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "topic",
				Value: "",
				Usage: "当前的类型有 compile trustedsetup genproofs verify",
			},
			cli.StringFlag{
				Name: "geth",
				Value: "",
				Usage: "geth",
			},
			cli.StringFlag{
				Name: "wasm",
				Value: "",
				Usage: "wasm",
			},
			cli.StringFlag{
				Name: "path",
				Value: "",
				Usage: "Path",
			},
		},
		// 指定该命令要执行的函数
		Action: RunZero,
	}
	err := app.Run(os.Args)
	if  err !=nil {
		log.Fatal(err)
	}

}

// 执行函数的结构体
type Zerocli struct{
	Topic string // 函数名
	Path string  // 路径
	Wasm string  //
	Geth string  // Geth16算法
}

// 执行函数
func RunZero(c *cli.Context)  {
	zcli := Zerocli{
		Topic: c.String("topic"),
		Path: c.String("path"),
		Wasm: c.String("wasm"),
		Geth: c.String("geth"),
	}
	// 若是geth16算法，则修改路径
	if c.String("geth") == "geth"{
		zcli.Path = zcli.Path+"geth16\\"
	}
	switch zcli.Geth {
	case "geth":
		switch zcli.Topic {
		case "compile":
			CompileCircuit(&zcli)
		case "trustedsetup":
			Groth16TrustedSetup(&zcli)
		case "genproofs":
			Groth16GenerateProofs(&zcli)
		case "verify":
			Groth16VerifyProofs(&zcli)
			break
		}
	case "":
		switch zcli.Topic {
		case "compile":
			CompileCircuit(&zcli)
		case "trustedsetup":
			TrustedSetup(&zcli)
		case "genproofs":
			GenerateProofs(&zcli)
		case "verify":
			VerifyProofs(&zcli)
			break
		}
		break

	}

}

// 编译一个电路,命令行输入的参数在context.flagSet.args中
func CompileCircuit(zcli *Zerocli) error {
	//fmt.Println("命令行中的参数：",context.Args())
	// Args()为获取命令行输入参数集合，Get（0）为获取第一个参数
	circuitPath := zcli.Path

	wasmFlag := false
	if zcli.Wasm == "wasm" {
		wasmFlag = true
	}

	// read circuit file，获取一个file结构体的地址
	circuitFile, err := os.Open(circuitPath+"test.tx")
	panicErr(err)

	// parse circuit code，创建一个新的解析器
	parser := circuitcompiler.NewParser(bufio.NewReader(circuitFile))
	// 解析并返回编译后的Circuit电路
	circuit, err := parser.Parse()
	panicErr(err)
	// 输出电路
	fmt.Println("\ncircuit data:", circuit)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile(circuitPath+"privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(circuitPath+"publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)
	fmt.Println("\nwitness", w)

	// flat code to R1CS
	fmt.Println("\ngenerating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	alphas, betas, gammas, zx := snark.Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	ax, bx, cx, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := snark.Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	// assert.Equal(t, px, snark.Utils.PF.Mul(hx, zx))
	if !r1csqap.BigArraysEqual(px, snark.Utils.PF.Mul(hx, zx)) {
		panic(errors.New("px != hx*zx"))
	}

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := snark.Utils.PF.Sub(snark.Utils.PF.Mul(ax, bx), cx)
	// assert.Equal(t, abc, px)
	if !r1csqap.BigArraysEqual(abc, px) {
		panic(errors.New("abc != px"))
	}
	hz := snark.Utils.PF.Mul(hx, zx)
	if !r1csqap.BigArraysEqual(abc, hz) {
		panic(errors.New("abc != hz"))
	}
	// assert.Equal(t, abc, hz)

	div, rem := snark.Utils.PF.Div(px, zx)
	if !r1csqap.BigArraysEqual(hx, div) {
		panic(errors.New("hx != div"))
	}
	// assert.Equal(t, hx, div)
	// assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))
	for _, r := range rem {
		if !bytes.Equal(r.Bytes(), big.NewInt(int64(0)).Bytes()) {
			panic(errors.New("error:error:  px/zx rem not equal to zeros"))
		}
	}

	// store circuit to json
	jsonData, err := json.Marshal(circuit)
	panicErr(err)
	// store setup into file
	jsonFile, err := os.Create(circuitPath+"compiledcircuit.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Compiled Circuit data written to ", jsonFile.Name())

	if wasmFlag {
		circuitString := utils.CircuitToString(*circuit)
		jsonData, err := json.Marshal(circuitString)
		panicErr(err)
		// store setup into file
		jsonFile, err := os.Create(circuitPath+"compiledcircuitString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}

	// store px
	jsonData, err = json.Marshal(px)
	panicErr(err)
	// store setup into file
	jsonFile, err = os.Create(circuitPath+"px.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Px data written to ", jsonFile.Name())
	if wasmFlag {
		pxString := utils.ArrayBigIntToString(px)
		jsonData, err = json.Marshal(pxString)
		panicErr(err)
		// store setup into file
		jsonFile, err = os.Create(circuitPath+"pxString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}

	return nil
}

func TrustedSetup(zcli *Zerocli) error {
	wasmFlag := false
	if zcli.Wasm == "wasm" {
		wasmFlag = true
	}
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile(zcli.Path+"compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile(zcli.Path+"privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)

	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(circuit.R1CS.A, circuit.R1CS.B, circuit.R1CS.C)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	// calculate trusted setup
	setup, err := snark.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas)
	panicErr(err)
	fmt.Println("\nt:", setup.Toxic.T)

	// remove setup.Toxic
	var tsetup snark.Setup
	tsetup.Pk = setup.Pk
	tsetup.Vk = setup.Vk
	tsetup.Pk.G1T = setup.Pk.G1T

	// store setup to json
	jsonData, err := json.Marshal(tsetup)
	panicErr(err)
	// store setup into file
	jsonFile, err := os.Create(zcli.Path+"trustedsetup.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Trusted Setup data written to ", jsonFile.Name())
	if wasmFlag {
		tsetupString := utils.SetupToString(tsetup)
		jsonData, err := json.Marshal(tsetupString)
		panicErr(err)
		// store setup into file
		jsonFile, err := os.Create(zcli.Path+"trustedsetupString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}
	return nil
}

func GenerateProofs(zcli *Zerocli) error {

	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile(zcli.Path+"compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile(zcli.Path+"trustedsetup.json")
	panicErr(err)

	var trustedsetup snark.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile(zcli.Path+"privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate witness见证
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)
	fmt.Println("witness", w)

	// flat code to R1CS
	a := circuit.R1CS.A
	b := circuit.R1CS.B
	c := circuit.R1CS.C
	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)
	_, _, _, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := snark.Utils.PF.DivisorPolynomial(px, trustedsetup.Pk.Z)

	fmt.Println("输出电路：",circuit)
	fmt.Println(trustedsetup.Pk.G1T)
	fmt.Println(hx)
	fmt.Println("输出witness：",w)
	proof, err := snark.GenerateProofs(circuit, trustedsetup.Pk, w, px)
	panicErr(err)

	fmt.Println("\n proofs:")
	fmt.Println(proof)

	// store proofs to json
	jsonData, err := json.Marshal(proof)
	panicErr(err)
	// store proof into file
	jsonFile, err := os.Create(zcli.Path+"proofs.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Proofs data written to ：", jsonFile.Name())
	return nil
}

func VerifyProofs(zcli *Zerocli) error {
	// open proofs.json
	proofsFile, err := ioutil.ReadFile(zcli.Path+"proofs.json")
	panicErr(err)
	var proof snark.Proof
	json.Unmarshal([]byte(string(proofsFile)), &proof)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile(zcli.Path+"trustedsetup.json")
	panicErr(err)
	var trustedsetup snark.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)
	var publicSignals []*big.Int
	err = json.Unmarshal([]byte(string(publicInputsFile)), &publicSignals)
	panicErr(err)

	verified := snark.VerifyProof(trustedsetup.Vk, proof, publicSignals, true)
	if !verified {
		fmt.Println("ERROR: proofs not verified")
	} else {
		fmt.Println("Proofs verified")
	}
	return nil
}


// 针对电路生成证明密钥和验证密钥
func Groth16TrustedSetup(zcli *Zerocli) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile(zcli.Path+"compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile(zcli.Path+"privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)

	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(circuit.R1CS.A, circuit.R1CS.B, circuit.R1CS.C)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	// calculate trusted setup
	setup, err := groth16.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas)
	panicErr(err)
	fmt.Println("\nt:", setup.Toxic.T)

	// remove setup.Toxic
	var tsetup groth16.Setup
	tsetup.Pk = setup.Pk
	tsetup.Vk = setup.Vk

	// store setup to json
	jsonData, err := json.Marshal(tsetup)
	panicErr(err)
	// store setup into file
	jsonFile, err := os.Create(zcli.Path+"trustedsetup.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Trusted Setup data written to ", jsonFile.Name())
	return nil
}

// 在给定witness/statement的情况下生成证明
func Groth16GenerateProofs(zcli *Zerocli) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile(zcli.Path+"compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile(zcli.Path+"trustedsetup.json")
	panicErr(err)
	var trustedsetup groth16.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile(zcli.Path+"privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)
	fmt.Println("witness", w)

	// flat code to R1CS
	a := circuit.R1CS.A
	b := circuit.R1CS.B
	c := circuit.R1CS.C
	// R1CS to QAP
	alphas, betas, gammas, _ := groth16.Utils.PF.R1CSToQAP(a, b, c)
	_, _, _, px := groth16.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := groth16.Utils.PF.DivisorPolynomial(px, trustedsetup.Pk.Z)

	fmt.Println(circuit)
	fmt.Println(trustedsetup.Pk.PowersTauDelta)
	fmt.Println(hx)
	fmt.Println(w)
	proof, err := groth16.GenerateProofs(circuit, trustedsetup.Pk, w, px)
	panicErr(err)

	fmt.Println("\n proofs:")
	fmt.Println(proof)

	// store proofs to json
	jsonData, err := json.Marshal(proof)
	panicErr(err)
	// store proof into file
	jsonFile, err := os.Create(zcli.Path+"proofs.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Proofs data written to ", jsonFile.Name())
	return nil
}

// 通过验证密钥验证证明是否正确
func Groth16VerifyProofs(zcli *Zerocli) error {
	// open proofs.json
	proofsFile, err := ioutil.ReadFile(zcli.Path+"proofs.json")
	panicErr(err)
	var proof groth16.Proof
	json.Unmarshal([]byte(string(proofsFile)), &proof)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile(zcli.Path+"trustedsetup.json")
	panicErr(err)
	var trustedsetup groth16.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile(zcli.Path+"publicInputs.json")
	panicErr(err)
	var publicSignals []*big.Int
	err = json.Unmarshal([]byte(string(publicInputsFile)), &publicSignals)
	panicErr(err)

	verified := groth16.VerifyProof(trustedsetup.Vk, proof, publicSignals, true)
	if !verified {
		fmt.Println("ERROR: proofs not verified")
	} else {
		fmt.Println("Proofs verified")
	}
	return nil
}
