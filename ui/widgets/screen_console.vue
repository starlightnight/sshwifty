<!--
// Sshwifty - A Web SSH client
//
// Copyright (C) 2019-2022 Ni Rui <ranqus@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

<template>
  <div class="screen-console">
    <div
      class="console-console"
      :style="'font-family: ' + typeface + ', inherit'"
    >
      <h2 style="display: none">Console</h2>

      <div class="console-loading">
        <div class="console-loading-frame">
          <div class="console-loading-icon"></div>
          <div class="console-loading-message">Initializing console ...</div>
        </div>
      </div>
    </div>

    <!--
      Tell you this: the background transparent below is probably the most
      important transparent setting in the entire application. Make sure user
      can see through it so they can operate the console while keep the toolbar
      open.
    -->
    <div
      v-if="toolbar"
      class="console-toolbar"
      :style="'background-color: ' + control.activeColor() + 'ee'"
    >
      <h2 style="display: none">Tool bar</h2>

      <div class="console-toolbar-group console-toolbar-group-left">
        <div class="console-toolbar-item">
          <h3 class="tb-title">Text size</h3>

          <ul class="lst-nostyle">
            <li>
              <a class="tb-item" href="javascript:;" @click="fontSizeUp">
                <span
                  class="
                    tb-key-icon tb-key-resize-icon
                    icon icon-keyboardkey1 icon-iconed-bottom1
                  "
                >
                  <i>+</i>
                  Increase
                </span>
              </a>
            </li>
            <li>
              <a class="tb-item" href="javascript:;" @click="fontSizeDown">
                <span
                  class="
                    tb-key-icon tb-key-resize-icon
                    icon icon-keyboardkey1 icon-iconed-bottom1
                  "
                >
                  <i>-</i>
                  Decrease
                </span>
              </a>
            </li>
          </ul>
        </div>
      </div>

      <div class="console-toolbar-group console-toolbar-group-main">
        <div
          v-for="(keyType, keyTypeIdx) in screenKeys"
          :key="keyTypeIdx"
          class="console-toolbar-item"
        >
          <h3 class="tb-title">{{ keyType.title }}</h3>

          <ul class="hlst lst-nostyle">
            <li v-for="(key, keyIdx) in keyType.keys" :key="keyIdx">
              <a
                class="tb-item"
                href="javascript:;"
                @click="sendSpecialKey(key[1])"
                v-html="$options.filters.specialKeyHTML(key[0])"
              ></a>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import FontFaceObserver from "fontfaceobserver";
import { Terminal } from "xterm";
import { WebLinksAddon } from "xterm-addon-web-links";
import { FitAddon } from "xterm-addon-fit";
import { isNumber } from "../commands/common.js";
import { consoleScreenKeys } from "./screen_console_keys.js";

import "./screen_console.css";
import "xterm/css/xterm.css";

const termTypeFace = "Hack";
const termFallbackTypeFace = "monospace";
const termTypeFaceLoadTimeout = 3000;
const termTypeFaceLoadError =
  'Remote font "' +
  termTypeFace +
  '" is unavailable, using "' +
  termFallbackTypeFace +
  '" instead until the remote font is loaded';
const termDefaultFontSize = 16;
const termMinFontSize = 8;
const termMaxFontSize = 36;

class Term {
  constructor(control) {
    const resizeDelayInterval = 500;

    this.control = control;
    this.closed = false;
    this.fontSize = termDefaultFontSize;
    this.term = new Terminal({
      allowTransparency: false,
      cursorBlink: true,
      cursorStyle: "block",
      fontFamily: termTypeFace + ", " + termFallbackTypeFace,
      fontSize: this.fontSize,
      logLevel: process.env.NODE_ENV === "development" ? "info" : "off",
    });
    this.fit = new FitAddon();

    this.term.loadAddon(this.fit);
    this.term.loadAddon(new WebLinksAddon());

    this.term.setOption("theme", {
      background: this.control.activeColor(),
    });

    this.term.onData((data) => {
      if (this.closed) {
        return;
      }

      this.control.send(data);
    });

    this.term.onBinary((data) => {
      if (this.closed) {
        return;
      }

      this.control.sendBinary(data);
    });

    this.term.onKey((ev) => {
      if (this.closed) {
        return;
      }

      if (!this.control.echo()) {
        return;
      }

      const printable =
        !ev.domEvent.altKey &&
        !ev.domEvent.altGraphKey &&
        !ev.domEvent.ctrlKey &&
        !ev.domEvent.metaKey;

      switch (ev.domEvent.key) {
        case "Enter":
          ev.domEvent.preventDefault();
          this.writeStr("\r\n");
          break;

        case "Backspace":
          ev.domEvent.preventDefault();
          this.writeStr("\b \b");
          break;

        default:
          if (printable) {
            ev.domEvent.preventDefault();
            this.writeStr(ev.key);
          }
      }
    });

    this.term.attachCustomKeyEventHandler(async (ev) => {
      if (this.closed) {
        return true;
      }

      if (
        ev.type == "keyup" &&
        ((ev.key.toLowerCase() === "v" && ev.shiftKey && ev.ctrlKey) ||
          (ev.key === "Insert" && ev.shiftKey))
      ) {
        try {
          let text = await window.navigator.clipboard.readText();

          this.writeEchoStr(text);
        } catch (e) {
          alert(
            "Unable to paste: " +
              e +
              ". Please try again without using the Control+Shift+V / " +
              "Shift+Insert hot key"
          );
        }
        return false;
      }

      if (
        ev.type == "keyup" &&
        ((ev.key.toLowerCase() === "c" && ev.shiftKey && ev.ctrlKey) ||
          (ev.key === "Insert" && ev.ctrlKey))
      ) {
        try {
          window.navigator.clipboard.writeText(this.term.getSelection());
        } catch (e) {
          alert("Unable to copy: " + e);
        }
        return false;
      }

      return true;
    });

    let resizeDelay = null,
      oldRows = 0,
      oldCols = 0;

    this.term.onResize((dim) => {
      if (this.closed) {
        return;
      }

      if (dim.cols === oldCols && dim.rows === oldRows) {
        return;
      }

      oldRows = dim.rows;
      oldCols = dim.cols;

      if (resizeDelay !== null) {
        clearTimeout(resizeDelay);
        resizeDelay = null;
      }

      resizeDelay = setTimeout(() => {
        resizeDelay = null;

        if (!isNumber(dim.cols) || !isNumber(dim.rows)) {
          return;
        }

        if (!dim.cols || !dim.rows) {
          return;
        }

        this.control.resize({
          rows: dim.rows,
          cols: dim.cols,
        });
      }, resizeDelayInterval);
    });
  }

  init(root, callbacks) {
    if (this.closed) {
      return;
    }

    this.term.open(root);

    this.term.textarea.addEventListener("focus", callbacks.focus);
    this.term.textarea.addEventListener("blur", callbacks.blur);

    this.refit();
  }

  dispatch(event) {
    if (this.closed) {
      return;
    }

    try {
      this.term.textarea.dispatchEvent(event);
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  writeEchoStr(d) {
    if (this.closed) {
      return;
    }

    this.control.send(d);

    if (!this.control.echo()) {
      return;
    }

    this.writeStr(d);
  }

  writeStr(d) {
    if (this.closed) {
      return;
    }

    try {
      this.term.write(d);
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  write(d) {
    if (this.closed) {
      return;
    }

    try {
      this.term.write(d);
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  setFont(value) {
    if (this.closed) {
      return;
    }

    this.term.setOption("fontFamily", value);
    this.refit();
  }

  fontSizeUp() {
    if (this.closed) {
      return;
    }

    if (this.fontSize >= termMaxFontSize) {
      return;
    }

    this.fontSize += 2;
    this.term.setOption("fontSize", this.fontSize);
    this.refit();
  }

  fontSizeDown() {
    if (this.closed) {
      return;
    }

    if (this.fontSize <= termMinFontSize) {
      return;
    }

    this.fontSize -= 2;
    this.term.setOption("fontSize", this.fontSize);
    this.refit();
  }

  focus() {
    if (this.closed) {
      return;
    }

    try {
      this.term.focus();
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  blur() {
    if (this.closed) {
      return;
    }

    try {
      this.term.blur();
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  refit() {
    if (this.closed) {
      return;
    }

    try {
      this.fit.fit();
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }

  destroyed() {
    return this.closed;
  }

  destroy() {
    if (this.closed) {
      return;
    }

    this.closed = true;

    try {
      this.term.dispose();
    } catch (e) {
      process.env.NODE_ENV === "development" && console.trace(e);
    }
  }
}

// So it turns out, display: none + xterm.js == trouble, so I changed this
// to a visibility + position: absolute appoarch. Problem resolved, and I
// like to keep it that way.

export default {
  filters: {
    specialKeyHTML(key) {
      const head = '<span class="tb-key-icon icon icon-keyboardkey1">',
        tail = "</span>";

      return head + key.split("+").join(tail + "+" + head) + tail;
    },
  },
  props: {
    active: {
      type: Boolean,
      default: false,
    },
    control: {
      type: Object,
      default: () => null,
    },
    change: {
      type: Object,
      default: () => null,
    },
    toolbar: {
      type: Boolean,
      default: false,
    },
    viewPort: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      screenKeys: consoleScreenKeys,
      term: new Term(this.control),
      typeface: termTypeFace,
      runner: null,
      eventHandlers: {
        keydown: null,
        keyup: null,
      },
    };
  },
  watch: {
    active(newVal, oldVal) {
      this.triggerActive(newVal);
    },
    change: {
      handler() {
        if (!this.active) {
          return;
        }

        this.fit();
      },
      deep: true,
    },
    viewPort: {
      handler() {
        if (!this.active) {
          return;
        }

        this.fit();
      },
      deep: true,
    },
  },
  async mounted() {
    await this.init();
  },
  beforeDestroy() {
    this.deinit();
  },
  methods: {
    loadRemoteFont(typeface, timeout) {
      return Promise.all([
        new FontFaceObserver(typeface).load(null, timeout),
        new FontFaceObserver(typeface, {
          weight: "bold",
        }).load(null, timeout),
      ]);
    },
    async retryLoadRemoteFont(typeface, timeout, onSuccess) {
      const self = this;

      for (;;) {
        if (self.term.destroyed()) {
          return;
        }

        try {
          onSuccess(await self.loadRemoteFont(typeface, timeout));

          return;
        } catch (e) {
          // Retry
        }
      }
    },
    async openTerm(root, callbacks) {
      const self = this;

      try {
        await self.loadRemoteFont(termTypeFace, termTypeFaceLoadTimeout);

        if (self.term.destroyed()) {
          return;
        }

        root.innerHTML = "";

        self.term.init(root, callbacks);

        return;
      } catch (e) {
        // Ignore
      }

      if (self.term.destroyed()) {
        return;
      }

      root.innerHTML = "";

      callbacks.warn(termTypeFaceLoadError, false);

      self.term.setFont(termFallbackTypeFace);
      self.term.init(root, callbacks);

      self.retryLoadRemoteFont(termTypeFace, termTypeFaceLoadTimeout, () => {
        if (self.term.destroyed()) {
          return;
        }

        self.term.setFont(termTypeFace);

        callbacks.warn(termTypeFaceLoadError, true);
      });
    },
    triggerActive(active) {
      active ? this.activate() : this.deactivate();
    },
    async init() {
      let self = this;

      self.eventHandlers = {
        keyup: (e) => self.localKeypress(e),
        keydown: (e) => self.localKeypress(e),
      };

      await self.openTerm(
        self.$el.getElementsByClassName("console-console")[0],
        {
          focus(e) {
            document.addEventListener("keyup", self.eventHandlers.keyup);
            document.addEventListener("keydown", self.eventHandlers.keydown);
          },
          blur(e) {
            document.removeEventListener("keyup", self.eventHandlers.keyup);
            document.removeEventListener("keydown", self.eventHandlers.keydown);
          },
          warn(msg, toDismiss) {
            self.$emit("warning", {
              text: msg,
              toDismiss: toDismiss,
            });
          },
          info(msg, toDismiss) {
            self.$emit("info", {
              text: msg,
              toDismiss: toDismiss,
            });
          },
        }
      );

      if (self.term.destroyed()) {
        return;
      }

      self.triggerActive(this.active);
      self.runRunner();
    },
    async deinit() {
      await this.closeRunner();
      await this.deactivate();
      this.term.destroy();
    },
    fit() {
      this.term.refit();
    },
    localKeypress(e) {
      if (!e.altKey && !e.shiftKey && !e.ctrlKey) {
        return;
      }

      e.preventDefault();
    },
    activate() {
      this.term.focus();
      this.fit();
    },
    async deactivate() {
      this.term.blur();
    },
    runRunner() {
      if (this.runner !== null) {
        return;
      }

      let self = this;

      this.runner = (async () => {
        try {
          for (;;) {
            if (self.term.destroyed()) {
              break;
            }

            self.term.write(await this.control.receive());

            self.$emit("updated");
          }
        } catch (e) {
          self.$emit("stopped", e);
        }
      })();
    },
    async closeRunner() {
      if (this.runner === null) {
        return;
      }

      let runner = this.runner;
      this.runner = null;

      await runner;
    },
    sendSpecialKey(key) {
      if (!this.term) {
        return;
      }

      this.term.dispatch(new KeyboardEvent("keydown", key));
      this.term.dispatch(new KeyboardEvent("keyup", key));
    },
    fontSizeUp() {
      this.term.fontSizeUp();
    },
    fontSizeDown() {
      this.term.fontSizeDown();
    },
  },
};
</script>
