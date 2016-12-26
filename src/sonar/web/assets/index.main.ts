/// <reference path="lib/xhr.ts" />

module app {
	interface Res<T> {
		ok: boolean;
		data: T;
	}

	interface Summary {
		avg: number;
		stddev: number;
		count: number;
		max: number;
		min: number;
		data?: number[];
	}

	interface Current extends Summary {
		ip: string;
		name: string;
		time: string;
	}

	interface Hour extends Summary {
		time: string;
	}

	interface Hourly {
		ip: string;
		name: string;
		hours: Hour[];
	}

	function toMs(ns: number) : number {
		return ns / 1e6;
	}

	k4.get('/api/v1/current')
		.onSuccess((json: string) => {
			var data = <Res<Current[]>>JSON.parse(json);
			var cnt = document.querySelector('#cnt'),
				el = cnt.appendChild(document.createElement('div'));
			data.data.forEach((c: Current) => {
				var div = document.createElement('div');
				div.textContent = toMs(c.avg).toFixed(2);
				el.appendChild(div);
			});
		});

	k4.get('/api/v1/hourly')
		.onSuccess((json: string) => {
			var data = <Res<Hourly[]>>JSON.parse(json);
			var cnt = document.querySelector('#cnt'),
				el = cnt.appendChild(document.createElement('div'));
			data.data.forEach((h) => {
				var v = h.hours.map((x) => {
					return toMs(x.avg);
				});

				var div = document.createElement('div');
				div.textContent = h.ip + ":" + JSON.stringify(v);
				el.appendChild(div);
			});
		});
}