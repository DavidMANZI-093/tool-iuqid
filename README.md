## tool-iquid
>
> *Automated recovery for the stubborn router.*

---

### 📖 The Story

#### The 7-Minute Itch

This goes back to November 2025. We had this Home WiFi for almost a year—juicy, fast fiber connection that could handle anything. But it had one fatal flaw: **fragility**.

A simple power cut, a little wind, some rain, or just the chaotic entropy of the universe would take the network down. The router would stay on, but the connection would vanish. The only fix? **Restart it.**

But this wasn't a quick "off and on again." One restart takes about **7 minutes** for full configuration and LED stabilization. And once it's back? "Internet connection not available."
It became a ritual. Restart. Wait 7 minutes. Fail. Restart. Wait 7 minutes. Fail. Sometimes it took 3 tries. Sometimes 20. Sometimes it took *all day*.

#### The First Crusade (Failure)

Being a "True Man of Culture," I couldn't accept playing these games. You give a portion of your life to a problem, but you fix it *once and for all*.

I pulled out `wget`, extracted the router's web content, and started inspecting `.cgi` and `.js` files. I had the admin credentials, but the router wasn't using simple Basic Auth. It was some complex JavaScript encryption scheme.
I spent a whole afternoon writing spaghetti JS code, raw-dogging the implementation, hoping to script something for my home server.
I burned out. I couldn't see how my code interacted with the router's mini-server. I was blindly poking a black box. I went to bed defeated.

#### The Sunday Morning Redemption

Fast forward to February 2026. A week and a half of constant drops. Sunday morning, 6 AM. I was fed up.

The router is across the office, so I couldn't even sit comfortably. I dragged an extension cord across the room and lay on the floor, ethernet cable plugged directly into the router, determined to end this.

This time, I chose **Go**. Clean, organized, but most importantly, a native binary.
I enlisted an AI agent (Antigravity) to help reverse-engineer the encryption protocol that beat me months ago. We found a dual-layer scheme using `sjcl` (AES) and `jsencrypt` (RSA). The router generated ephemeral keys on every page load, requiring a perfect crypto-handshake just to say "reboot."

It fought us using `base64URL` encoding quirks and hidden form tokens. It felt impossible halfway through. But having been there before, I knew it *could* be done. We pushed.
10 hours later... success. A dry run authenticated. The crawler found the reboot URL. The payload triggered.

#### The Result

That Sunday night, I refactored the experimental mess into **tool-iquid**.
It's a minimal, lightweight, robust Go service. It sits on my home server, watching the connection like a hawk. When the internet drops, it steps in, handles the cryptographic handshake, and reboots the router automatically.

I had fun. I had thrills. I had twists. And now, I have a fairly reliable internet connection.

---

### 🛠️ What It Does

**tool-iquid** is a daemon that:

1. **Monitors** your internet connection (checking `8.8.8.8:53`).
2. **Verifies** it's connected to the correct SSID (to avoid rebooting the wrong network).
3. **Authenticates** with the router using a reverse-engineered AES+RSA encryption scheme.
4. **Triggers** a reboot if connectivity is lost.
5. **Waits** through the 7-minute initialization period before checking again.

### 🚀 Getting Started

#### Installation

```bash
git clone https://github.com/DavidMANZI-093/tool-iquid.git
cd tool-iquid
go build -o tool-iquid ./cmd/tool-iquid
```

#### Usage

Run it with your config:

```bash
./tool-iquid -config config.json
```

For full details on configuration, systemd setup, and command-line flags, see the **[User Manual](docs/manual.md)**.
For the nerdy details on how we cracked the encryption, read **[The Reverse Engineering Docs](docs/encryption_reverse_engineering.md)**.

### 📜 License

Unlicense. A piece of my life given to the world.

---
*"A True Man of Culture doesn't play such games."*
