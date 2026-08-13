import { createRequire } from "module";
import { dirname, join } from "path";
import { fileURLToPath } from "url";
import WebSocket from "ws";

const require = createRequire(import.meta.url);
const here = dirname(fileURLToPath(import.meta.url));
const VERSION = (require(join(here, "..", "package.json")) as { version: string }).version;

const PORT = Number(process.env.NEXUS_WS_PORT ?? 49152);
const SOURCE = "S-D";
const SERVICE_ID = "s-d-integrations";
const HELLO_INTERVAL_MS = 5000;

const url = `ws://127.0.0.1:${PORT}/`;
const ws = new WebSocket(url);

function hello(): string {
  return JSON.stringify({
    jsonrpc: "2.0",
    method: "event.system.hello",
    params: {
      source: SOURCE,
      protocol_version: 1,
      service_id: SERVICE_ID,
      version: VERSION,
      ts: new Date().toISOString(),
    },
  });
}

ws.on("open", () => {
  console.log(`[${SERVICE_ID}] verbunden mit ${url}`);
  sendHello();
  setInterval(sendHello, HELLO_INTERVAL_MS);
});

ws.on("message", () => {});

ws.on("error", (err) => {
  console.error(`[${SERVICE_ID}] Fehler: ${err.message}`);
});

function sendHello(): void {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(hello());
  }
}
