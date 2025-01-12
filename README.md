#TODO

# dShout

Store encrypted messages for one or more receivers on chain

---

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
Uncheck **Restrictive Mode** and set **Global Permissions** to `Apply`. 
Go back and turn **Web Sockets (WS)** on.
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
