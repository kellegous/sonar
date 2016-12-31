/// <reference path="lib/xhr.ts" />
/// <reference path="lib/signal.ts" />

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

	interface Report {
		ip: string;
		name: string;
		currently: Summary;
		hourly: Hour[];
	}

	function toValues<T>(obj: {[key: string]: T}): T[] {
		var vals: T[] = [];
		for (var key in obj) {
			vals.push(obj[key]);
		}
		return vals;
	};

	function start(s : Q.Signal, period: number) {
		k4.mget('/api/v1/current', '/api/v1/hourly')
			.onSuccess((json: string[]) => {
				var curr = <Res<Current[]>>JSON.parse(json[0]),
					hrly = <Res<Hourly[]>>JSON.parse(json[1]),
					data: {[key: string]: Report} = {};

				curr.data.forEach((c) => {
					data[c.ip] = {
						ip: c.ip,
						name: c.name,
						currently: c,
						hourly: []
					};
				});

				hrly.data.forEach((h) => {
					data[h.ip].hourly = h.hours;
				});

				s.raise(toValues(data));
			});
	}

	// k4.get('/api/v1/current')
	// 	.onSuccess((json: string) => {
	// 		var data = <Res<Current[]>>JSON.parse(json);
	// 		var cnt = document.querySelector('#cnt'),
	// 			el = cnt.appendChild(document.createElement('div'));
	// 		data.data.forEach((c: Current) => {
	// 			var div = document.createElement('div');
	// 			div.textContent = toMs(c.avg).toFixed(2);
	// 			el.appendChild(div);
	// 		});
	// 	});

	// k4.get('/api/v1/hourly')
	// 	.onSuccess((json: string) => {
	// 		var data = <Res<Hourly[]>>JSON.parse(json);
	// 		var cnt = document.querySelector('#cnt'),
	// 			el = cnt.appendChild(document.createElement('div'));
	// 		data.data.forEach((h) => {
	// 			var v = h.hours.map((x) => {
	// 				return toMs(x.avg);
	// 			});

	// 			var div = document.createElement('div');
	// 			div.textContent = h.ip + ":" + JSON.stringify(v);
	// 			el.appendChild(div);
	// 		});
	// 	});

	var didLoad = new Q.Signal();
	didLoad.tap((report: Report) => {
		console.log(report);
	});
	start(didLoad, 60*1000);
}