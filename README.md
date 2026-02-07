# go-ink

A native Go runtime for the **Ink** narrative scripting language.

`go-ink` allows you to load and play stories written in [Inkle's Ink language](https://github.com/inkle/ink) directly within your Go applications. It is designed to be idiomatic, fast, and fully compatible with the standard Ink JSON format.

## 📖 Origin & Attribution

This project is a derivative work, explicitly based on the architectural foundations of **[blade-ink](https://github.com/bladecoder/blade-ink)**.

We owe a significant debt of gratitude to **[bladecoder](https://github.com/bladecoder)** and the `blade-ink` contributors. Their robust Java implementation served as the primary reference for this port, ensuring that `go-ink` benefits from the stability and logic verifications of its predecessor.

We also acknowledge **[Inkle](https://www.inklestudios.com/)** for creating the Ink language and the original C# runtime that started it all.

### 📄 Why a Go Port?

The Ink language is an incredibly powerful tool for non-linear storytelling, but the official runtime ecosystem often forces a binary choice: build a heavy C# project in Unity, or build a web-based game in JavaScript.

I built ink-go to strike a balance between those two extremes.
1. The "Middle Path" (Unity vs. Ren'Py)

    Not Unity: Unity is a massive engine with significant overhead. For many 2D or text-heavy games, bringing in the entire Unity/C# ecosystem just to parse Ink stories feels like overkill.

    Not Ren'Py: While Ren'Py is great for Visual Novels, it locks you into its specific Python-based architecture. I wanted the narrative freedom of Ink without being constrained to the visual novel genre or the Ren'Py engine limitations.

2. Control & Familiarity

    Go-Native: I wanted to write game logic in a language I enjoy and am familiar with—Go.

    Engine Agnostic: By porting the full runtime to pure Go, I can integrate Ink into any Go-based game framework (like Ebitengine) completely natively. There is no CGO bridging, no embedded C# VM, and no "black box" engine logic.

3. Lightweight & Portable

    Single Binary: The result is a high-performance narrative engine that compiles into a single, static binary. It runs anywhere Go runs, with zero dependencies on external runtimes or frameworks.

## ✅ Conformance

`go-ink` is built to be strictly compliant with the Ink language specification. It has been verified against the standard **Ink Conformance Test Suite**.

**Current Support:**
* **Flow Control:** Knots, Stitches, Diverts (`->`), and Tunnels.
* **Logic:** Full variable support (Global & Temporary), mathematical operations (`+`, `-`, `*`, `/`, `%`), and conditionals (`==`, `!=`, `>`, `<`).
* **Text Processing:** Correct handling of "Glue" (`<>`) and whitespace trimming.
* **Native Functions:** Built-in Ink functions are fully implemented.
* **JSON Parsing:** Recursive descent parser for standard `.ink.json` exports.

## 📦 Installation

```bash
go get [github.com/your-username/go-ink](https://github.com/your-username/go-ink)
```

## 🚀 Usage

Here is a minimal example of how to load a story and play it line-by-line:

```
package main

import (
	"fmt"
	"log"
	"os"

	"[github.com/your-username/go-ink/ink](https://github.com/your-username/go-ink/ink)"
)

func main() {
	// 1. Read the compiled JSON file
	jsonBytes, err := os.ReadFile("story.ink.json")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize the Story
	story, err := ink.NewStory(string(jsonBytes))
	if err != nil {
		log.Fatal(err)
	}

	// 3. The Game Loop
	for story.CanContinue() {
		// Continue() calculates the next line of text
		text, err := story.Continue()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(text)
	}
    
    // 4. Check for Choices
    if len(story.CurrentChoices()) > 0 {
        // Handle choices...
    }
}
```

## ⚖️ License

This project is released under the MIT License, maintaining the same licensing terms as the original blade-ink and ink runtimes to ensure open ecosystem compatibility.

## 🤖 Development Methodology

This project is a **human-architected, AI-assisted port**.

The implementation was developed using **Google Gemini Code Assist** & **Github Copilot**, functioning as a pair programmer to translate the reference Java architecture (`blade-ink`) into idiomatic Go.

**The Protocol:**
1.  **Architecture First:** All critical structural decisions (Pointer semantics, Interface design, Stack architecture) were defined by human engineering oversight before implementation.
2.  **Strict Verification:** No AI-generated code was committed without passing specific logic gates (Glue logic, Math precision, Variable scoping).
3.  **Conformance:** The final runtime is validated against the official **Ink Conformance Test Suite** to ensure 100% parity with the standard Ink engine.