const API_PREFIX = "https://192.168.1.150:8000/api"

export async function fetchFromAPI<T>(url: string, mock: T): Promise<T> {
    const response = await fetch(`${API_PREFIX}${url}`);

    if (!response.ok) {
        console.error(`Fetch error (${response.status}), using mock...`);
        return mock;
    } else {
        return await response.json(); // update state with fetched data
    }
}
