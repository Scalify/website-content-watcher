import {getIp} from 'shared';

await page.goto(vars.page);
const ip = await getIp(page);

logger.info(ip);
results.ip = ip;
