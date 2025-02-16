import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const slowResponseRate = new Rate('slow_responses');
const items = Array('t-shirt', 'cup', 'book', 'pen', 'powerbank', 'hoody', 'umbrella', 'socks', 'wallet', 'pink-hoody')
const coinsOnRegister = 1000
let registeredUsers = []

export const options = {
	scenarios: {
		ramping_rate: {
			executor: 'ramping-arrival-rate',
			preAllocatedVUs: 10,
			maxVUs: 3000,
			startRate: 0,
			timeUnit: '1s',
			// stages: [ // auth only
			// 	{ duration: '1m', target: 1000 },
			// 	{ duration: '3m', target: 1000 },
			// 	{ duration: '1m', target: 0 },
			// ],
			// stages: [ // auth + buy item
			// 	{ duration: '1m', target: 500 },
			// 	{ duration: '3m', target: 500 },
			// 	{ duration: '1m', target: 0 },
			// ],
			// stages: [ // auth + buy item + get info
			// 	{ duration: '1m', target: 333 },
			// 	{ duration: '3m', target: 333 },
			// 	{ duration: '1m', target: 0 },
			// ],
			stages: [ // auth + buy item + send coins + get info
				{ duration: '1m', target: 250 },
				{ duration: '3m', target: 250 },
				{ duration: '1m', target: 0 },
			],
		}
	},
	thresholds: {
		http_req_duration: ['p(95)<60'],
		slow_responses: ['rate<0.01'],
	}
};

function randomString(length) {
	let result = '';
	const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
	const charactersLength = characters.length;
	let counter = 0;
	while (counter < length) {
		result += characters.charAt(Math.floor(Math.random() * charactersLength));
		counter += 1;
	}
	return result;
}

export function authLoadTest() {
	const payload = {
		"username": randomString(10),
		"password": "abcdef12345",
	};

	const params = {
		headers: {
		'accept': 'application/json',
		'Content-Type': 'application/json',
		},
	};

	const response = http.post('http://avito-shop-service:8080/api/auth', JSON.stringify(payload), params);

	check(response, {
		'Status is 200': (r) => r.status === 200,
		'Response contains token': (r) => r.body.includes('token'),
		'Response time is less than 50ms': (r) => r.timings.duration < 50
	});

	if (response.timings.duration >= 50) {
		slowResponseRate.add(1);
		console.log(`Slow response detected: ${response.timings.duration}ms`);
	}
	if (response.status === 200) {
		registeredUsers.push(payload["username"])
	}

	return response.json().token
}

export function buyItemLoadTest(token) {
	const item = items[Math.floor(Math.random()*items.length)];

	const params = {
		headers: {
			'Authorization': `Bearer ${token}`,
			'accept': '*/*',
		},
	};

	const response = http.get(`http://avito-shop-service:8080/api/buy/${item}`, params);

	check(response, {
		'Buy item status is 200': (r) => r.status === 200,
		'Response time is less than 50ms': (r) => r.timings.duration < 50
	});

	if (response.timings.duration >= 50) {
		slowResponseRate.add(1);
		console.log(`Slow response detected: ${response.timings.duration}ms`);
	}
}

export function getInfoLoadTest(token) {
	const params = {
		headers: {
			'Authorization': `Bearer ${token}`,
			'accept': 'application/json',
		},
	};

	const response = http.get(`http://avito-shop-service:8080/api/info`, params);

	check(response, {
		'Get info status is 200': (r) => r.status === 200,
		'Response body includes coins': (r) => r.body.includes('coins'),
		'Response body includes inventory': (r) => r.body.includes('inventory'),
		'Response body includes coinHistory': (r) => r.body.includes('coinHistory'),
		'Response time is less than 50ms': (r) => r.timings.duration < 50
	});

	if (response.timings.duration >= 50) {
		slowResponseRate.add(1);
		console.log(`Slow response detected: ${response.timings.duration}ms`);
	}
}

export function sendCoinsLoadTest(token) {
	const toUser = registeredUsers[Math.floor(Math.random()*registeredUsers.length)];

	const payload = {
		"toUser": toUser,
		"amount": Math.floor(Math.random()*coinsOnRegister),
	};

	const params = {
		headers: {
			'Authorization': `Bearer ${token}`,
			'accept': '*/*',
			'Content-Type': 'application/json',
		},
	};

	const response = http.post('http://avito-shop-service:8080/api/sendCoin', JSON.stringify(payload), params);

	// 400 - user not found, fromUser == toUser, not enough coins
	check(response, {
		'Send coins status is 200 or 400': (r) => r.status === 200 || r.status === 400,
		'Response time is less than 50ms': (r) => r.timings.duration < 50
	});

	if (response.status === 500) {
		console.log(payload)
	}

	if (response.timings.duration >= 50) {
		slowResponseRate.add(1);
		console.log(`Slow response detected: ${response.timings.duration}ms`);
	}
}

export default function () {
	let token = authLoadTest();
	buyItemLoadTest(token)

	sendCoinsLoadTest(token)

	getInfoLoadTest(token)

	sleep(1);
}
