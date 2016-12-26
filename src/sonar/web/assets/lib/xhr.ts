module k4 {
	export interface Xhr<T> {
		onSuccess(cb: (data: T) => void) : Xhr<T>;
		onError(cb: (msg: string) => void) : Xhr<T>;
	}

	class XhrImpl<T> implements Xhr<T> {
		success: (data: any) => void;
		error: (msg: string) => void;

		onSuccess(cb: (data: T) => void) : Xhr<T> {
			this.success = cb;
			return this;
		}

		onError(cb: (msg: string) => void) : Xhr<T> {
			this.error = cb;
			return this;
		}

		doSuccess(data: T) {
			if (this.success) {
				this.success(data);
			}
		}

		doError(msg: string) {
			if (this.error) {
				this.error(msg);
			}
		}
	}

	export function get(url: string): Xhr<string> {
		var xhr = new XMLHttpRequest,
			impl = new XhrImpl<string>();
		xhr.onreadystatechange = () => {
			if (xhr.readyState != 4) {
				return;
			}

			if (xhr.status >= 400) {
				impl.doError("error");
			} else {
				impl.doSuccess(xhr.responseText);
			}
		};
		xhr.open('GET', url);
		xhr.send();
		return impl;
	}

	export function mget(...urls: string[]): Xhr<string[]> {
		var count = urls.length,
			res = [],
			hadErr = false,
			impl = new XhrImpl<string[]>();
		urls.forEach((url, i) => {
			get(url).onSuccess((data: string) => {
					res[i] = data;
					count--;
					if (count > 0) {
						return;
					}
					impl.doSuccess(res);
				}).onError((msg: string) => {
					if (hadErr) {
						return;
					}
					hadErr = true;
					impl.doError(msg);
				});
		});

		return impl;
	}
}