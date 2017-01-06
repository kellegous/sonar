module dom {
	export class El {
		constructor(public el: HTMLElement) {
		}

		set(el: HTMLElement) : El {
			this.el = el;
			return this;
		}

		rel() : HTMLElement {
			var el = this.el;
			this.el = null;
			return el;
		}

		setCss(css: {[prop: string]: string}) : El {
			var style = this.el.style;
			for (var p in css) {
				style.setProperty(p, css[p], '');
			}
			return this;
		}

		setAttrs(attrs: {[name: string]: any}) : El {
			var el = this.el;
			for (var name in attrs) {
				el.setAttribute(name, '' + attrs[name]);
			}
			return this;
		}

		addClass(c: string) : El {
			this.el.classList.add(c);
			return this;
		}

		appendTo(el: HTMLElement) : El {
			el.appendChild(this.el);
			return this;
		}

		setText(text: string) : El {
			this.el.textContent = text;
			return this;
		}

		do(fn: (el: El) => void) : El {
			fn(this);
			return this;
		}
	}

	export function of(el: Element) : El {
		return new El(<HTMLElement>el);
	}

	export function create(name: string, ns?: string) : El {
		return of(ns !== undefined
			? document.createElementNS(ns, name)
			: document.createElement(name));
	}
}