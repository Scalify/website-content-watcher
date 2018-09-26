export async function getIp(page) {
    const text = await page.evaluate(() => document.querySelector('body').textContent);
    return text.split(":")[1];
}
