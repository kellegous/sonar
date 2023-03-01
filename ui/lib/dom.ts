type NativeEl = HTMLElement | SVGElement;

export class El<T extends NativeEl = NativeEl> {
	constructor(public el: T) {
	}

	set(el: T): El {
		this.el = el;
		return this;
	}

	rel(): T {
		return this.el;
	}

	setCss(css: { [prop: string]: string }): El {
		var style = this.el.style;
		for (var p in css) {
			style.setProperty(p, css[p], '');
		}
		return this;
	}

	setAttrs(attrs: { [name: string]: any }): El {
		var el = this.el;
		for (var name in attrs) {
			el.setAttribute(name, '' + attrs[name]);
		}
		return this;
	}

	addClass(c: string): El {
		this.el.classList.add(c);
		return this;
	}

	appendTo(el: NativeEl): El {
		el.appendChild(this.el);
		return this;
	}

	setText(text: string): El {
		this.el.textContent = text;
		return this;
	}

	do(fn: (el: El) => void): El {
		fn(this);
		return this;
	}

	static of<T extends NativeEl = NativeEl>(el: T): El {
		return new El(el);
	}

	static create<T extends NativeEl>(name: string, ns?: string): El {
		const el = (ns !== undefined
			? document.createElementNS(ns, name)
			: document.createElement(name)) as T;
		return El.of<T>(el);
	}
}
