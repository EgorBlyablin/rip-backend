
export interface Turbine {
    id: number;
    title: string;
    description: string;
    image: string | null;
    power: number;
    height: number;
}

export interface GenerationRequestDraftStats {
    id: number;
    turbinesCount: number;
}
