import { fetchFromAPI } from "./fetch"
import type { Turbine } from "./interfaces"
import { turbinesMock } from "./mock"


export const fetchTurbines = (titleFilter: string) => fetchFromAPI<Turbine[]>(
    titleFilter
        ? `/turbines/?turbineTitle=${encodeURIComponent(titleFilter)}`
        : `/turbines/`,
    turbinesMock.filter(turbine => turbine.title.toLowerCase().includes(titleFilter.toLowerCase()))
)

export const fetchTurbine = (id: number) => fetchFromAPI<Turbine>(
    `/turbines/${id}`,
    turbinesMock.find(turbine => turbine.id === id) || null
)