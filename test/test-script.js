let args = ""
process.argv.forEach(arg => {
    args += " " + arg
})

console.log("steps-npm test drive running custom npm script...")
console.log("process.argv value: " + args)
console.log("test script done!")