const API_PREFIX = "/api"

export async function fetchFromAPI<T>(url: string, mock: T | null): Promise<T | null> {
    const response = await fetch(`${API_PREFIX}${url}`);
        
    if (!response.ok) {
        console.error(`Fetch error (${response.status}), using mock...`);
        return mock;
    } else {
        return await response.json(); // update state with fetched data
    }
}
