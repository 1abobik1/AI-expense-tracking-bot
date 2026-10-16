// Вместо build/worker.mjs: один экземпляр Go на изолят, на запрос обновляются только env и ctx.
import "./build/wasm_exec.js";
import { connect } from "cloudflare:sockets";
import mod from "./build/app.wasm";

globalThis.tryCatch = (fn) => {
  try {
    return { result: fn() };
  } catch (e) {
    return { error: e };
  }
};

const context = { env: undefined, ctx: undefined, connect, binding: {} };
let started;

function start() {
  const go = new Go();
  let ready;
  const readyPromise = new Promise((resolve) => (ready = resolve));
  const instance = new WebAssembly.Instance(mod, {
    ...go.importObject,
    workers: { ready: () => ready() },
  });
  go.run(instance, context); // не ждём: main живёт, пока жив изолят
  return readyPromise;
}

export default {
  async fetch(req, env, ctx) {
    context.env = env;
    context.ctx = ctx;
    started ??= start();
    await started;
    return context.binding.handleRequest(req);
  },
};
