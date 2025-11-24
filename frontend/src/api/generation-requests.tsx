import { fetchFromAPI } from "./fetch"
import type { GenerationRequestDraftStats } from "./interfaces"


export const fetchRequestStats = () => fetchFromAPI<GenerationRequestDraftStats>(
    "/generation-requests/draft",
    {
        GenerationRequestId: 0,
        TurbinesCount: 0
    }
)
