import http from "k6/http";
import { check, sleep } from "k6";

const testURL = 'http://host.docker.internal:8092';

export let options = {
    thresholds: {
      checks: ['rate==1'],
    },
};

export default function (data) {
    var url = testURL + '/tag';
    var payload = 'Mama su kasa kasa smėlį. Žalia, balta, žalia, balta, huh';
    var params = {
        headers: {
            'Content-Type': 'plain/text',
        },
    };
    let res = http.post(url, payload, params);
    check(res, {
        "status was 200": (r) => r.status == 200,
        "transaction time OK": (r) => r.timings.duration < 500
    });
    sleep(0.1);
}
