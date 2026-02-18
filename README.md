# Learn ETH Clients: Deep Dive into Geth

Welcome to the **Learn ETH Clients** repository. This project is a dedicated space for Ethereum enthusiasts, researchers, and developers who want to look "under the hood" of the blockchain. 

While many understand Ethereum at a high level, this repo aims to bridge the gap between theoretical protocol design and the actual engineering reality of the world's most widely used execution client: **Go-Ethereum (Geth)**.

---

## 🎯 Our Mission
Our goal is to deconstruct the Ethereum protocol by examining how it is implemented in code. We don't just ask *what* the protocol does; we ask *how* Geth executes it efficiently, securely, and at scale.

## 🔍 What We Explore
We break down the complex layers of the Ethereum stack, focusing on the Go implementation of:

* **The State & Storage:** Understanding Merkle Patricia Tries and how Geth manages the global state.
* **The EVM (Ethereum Virtual Machine):** A deep dive into the interpreter, opcodes, and gas mechanics within the `core/vm` directory.
* **The Networking Layer:** Exploring RLPx, devp2p, and how nodes discover and communicate with each other.
* **Transaction Processing:** The journey of a transaction from the mempool (TxPool) to being included in a block.
* **Consensus Engine Interface:** How Geth interacts with the Consensus Layer (Engine API) via the Beacon Chain.

## 🛠 Why Geth?
Go-Ethereum is the "gold standard" implementation of the Ethereum execution layer. By studying Geth, you are learning the production-grade patterns that secure billions of dollars in assets. Understanding this codebase is the fastest way to transition from a Web3 user to a Protocol Engineer.

## 📂 Repository Structure
* `/docs`: Conceptual breakdowns of Ethereum Improvement Proposals (EIPs).
* `/code-walkthroughs`: Annotated snippets of the Geth source code.
* `/diagrams`: Visual flows of data through a Geth node.
* `/exercises`: Small tasks to help you navigate and modify a local Geth build.

---

## 🚀 How to Use This Repo
1.  **Start with the Docs:** Read through our conceptual overviews to get the "Why."
2.  **Follow the Path:** We recommend starting with the `core/` directory in Geth, as it contains the heart of the state transition logic.
3.  **Contribute:** Found a specific function in Geth that is particularly elegant or confusing? Open a PR and add an explanation to our walkthroughs!

---

> "The best way to understand a protocol is to read the code that breathes life into it."
