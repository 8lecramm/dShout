# dShout

Store encrypted messages for one or more receivers on chain.
The message is encrypted only once, since all receivers use the same shared secret.

---

## Requirements

- Go
- Engram (dev branch)

## Installation

Follow these steps to set up the project on your local machine:

1. Clone the repository:

   ```sh
   git clone https://github.com/8lecramm/dShout.git
   ```

2. Navigate to the project directory:

   ```sh
   cd dShout
   ```

3. Download the dependencies:

   ```sh
   go mod tidy
   ```

4. Compile the project:

   ```sh
   go build
   ```

---

## Usage

1. Start Engram, switch to **Cyberdeck** and go to **Settings**.
Uncheck **Restrictive Mode**, go back and turn **Web Sockets (WS)** on.
---
dShout needs the following permissions:

- **AttemptEPOCH** (not used yet)
- **QueryKey** (mnemonics) to recover the private key (used for decrypting messages)
- **Transfer** (used for creating SC calls and deploying messages on chain)
---
2. Run the application:

   ```sh
   ./dShout
   ```

3. Accept permission requests

### Encrypt a message
- enter wallet address(es); one per line
- write  a message
- click on **Generate output** to create the ciphertext
- choose a ringsize and click on **Send to SC**

### Read messages
- click on **Check for messages**
- a popup tells you if there are messages
- click on **Read messages** to open the message window

---

## Smart Contract

**SCID**: a8ee7e571130342e0b7baa9052ccbfe3c1766cc454403721d2a357e7eda14894

```
Function Initialize() Uint64
 10 STORE("height", BLOCK_HEIGHT())
 20 STORE("prev", BLOCK_HEIGHT())
 30 RETURN 0 
End Function 

Function Store(data String) Uint64
 /*  189 chars = public key (66 chars) + encrypted shared keys (66 chars each) + 1 seperator (1 char) + encrypted message (at least 28 bytes/56 chars) */
 10  IF STRLEN(data) < 189 THEN GOTO 130
 20  DIM h as Uint64
 30  DIM ph as Uint64
 40  LET h = BLOCK_HEIGHT()
 50  LET ph = LOAD("height")
 60  IF h == ph THEN GOTO 100
 70  STORE("msg",data)
 80  STORE("prev", ph)
 90  GOTO 110
 100 STORE("msg",LOAD("msg")+"+"+data)
 110 STORE("height",h)
 120 RETURN 0
 130 RETURN 1
End Function
```

The SC makes use of Graviton snapshots, because every SC call overwrites the `msg` variable, but previous values are still accessible.

---

## Technical

- messaging is also possible with normal transactions, but the payload (message length) is limited.
- pruned nodes discard transactions. Messages before the pruning height are no longer available.
- Smart Contracts store keys and values in the SC Meta tree and are available on pruned nodes.
- neither the sender, nor the receiver(s) will be revealed. Add your own wallet address to the receiver list to keep track of outgoing messages. When sending the message, make sure to use a suited ringsize.
