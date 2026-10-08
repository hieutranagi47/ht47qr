const isBrowser =
  typeof window !== "undefined" && typeof window.document !== "undefined";

// Node globals are optional in this browser-targeted module.
declare const process: { versions?: { node?: string } } | undefined;

const isNode =
  typeof process !== "undefined" &&
  process.versions != null &&
  process.versions.node != null;

const isWebWorker =
  typeof self === "object" &&
  self.constructor &&
  self.constructor.name === "DedicatedWorkerGlobalScope";

/**
 * @see https://github.com/jsdom/jsdom/releases/tag/12.0.0
 * @see https://github.com/jsdom/jsdom/issues/1537
 */
const isJsDom =
  (typeof window !== "undefined" && window.name === "nodejs") ||
  (typeof navigator !== "undefined" &&
    (navigator.userAgent.includes("Node.js") ||
      navigator.userAgent.includes("jsdom")));

declare var Deno: any; // should remove it: https://github.com/flexdinesh/browser-or-node/blob/master/src/index.js
const isDeno = typeof Deno !== "undefined" && typeof Deno.core !== "undefined";

export { isBrowser, isWebWorker, isNode, isJsDom, isDeno };
