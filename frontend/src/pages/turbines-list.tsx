import { useEffect, useState, type FC, type FormEvent } from "react";
import { Button, Col, Container, Form, InputGroup, Row } from "react-bootstrap";
import { useSearchParams } from "react-router";
import { fetchTurbines } from "../api/turbines";
import SearchIcon from "../assets/search.svg?react";
import { Breadcrumbs } from "../components/breadcrumbs";
import { TurbineCard } from "../components/turbine-card";
import { ROUTE_LABELS, ROUTES } from "../routes";
import type { Turbine } from "../api/interfaces";

export const TurbinesListPage: FC = () => {
    const [searchParams, setSearchParams] = useSearchParams();

    const generationRequest = { id: 0, turbinesCount: 0 };
    // Get initial search query from URL
    const titleFilter = searchParams.get("title-filter") || "";
    const [searchValue, setSearchValue] = useState(titleFilter);
    const [turbines, setTurbines] = useState<Turbine[]>([]);

    useEffect(() => {
        const fetchTurbinesWrapper = async () => {
            const turbinesData = await fetchTurbines(titleFilter);
            setTurbines(turbinesData || []);
        };

        fetchTurbinesWrapper();
    }, [titleFilter]);


    // Handle form submission for search
    const handleSearchSubmit = (e: FormEvent) => {
        e.preventDefault();

        const params = new URLSearchParams(searchParams);
        if (searchValue.trim()) {
            params.set("title-filter", searchValue.trim());
        } else {
            params.delete("title-filter");
        }
        setSearchParams(params);
    };

    return (
        <Container className="pb-5">
            <Breadcrumbs
                breadcrumbs={[
                    { label: ROUTE_LABELS.HOME, path: ROUTES.HOME },
                    { label: ROUTE_LABELS.TURBINES_LIST, path: ROUTES.TURBINES_LIST },
                ]}
            />
            <div
                style={{
                    display: "flex",
                    justifyContent: "space-between",
                    paddingBlock: 18,
                }}
            >
                <h1
                    style={{
                        textTransform: "uppercase",
                        fontWeight: "bold",
                        color: "#5B5B5B",
                        alignSelf: "center"
                    }}
                >
                    Ветрогенераторы
                    {titleFilter && <> (поиск: "{titleFilter}")</>} {/* Use titleFilter here too */}
                </h1>
                <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                    <Button
                        variant="dark"
                        style={{ textWrap: "nowrap" }}
                        disabled
                    >
                        Расчет ({generationRequest.turbinesCount})
                    </Button>
                    <Form onSubmit={handleSearchSubmit}>
                        <InputGroup>
                            <Form.Control
                                name="title-filter"
                                placeholder="Поиск"
                                value={searchValue}
                                onChange={({ currentTarget: { value } }) => setSearchValue(value)}
                                maxLength={20}
                            />
                            <Button type="submit" style={{ display: "flex", alignItems: "center", backgroundColor: "#5BA1D4", border: "none" }}>
                                <SearchIcon style={{ width: "1.2rem", height: "1.2rem" }} />
                            </Button>
                        </InputGroup>
                    </Form>
                </div>
            </div>
            <Row className="row-gap-4">
                {(turbines || []).map((turbine) => (
                    <Col key={turbine.id}>
                        <TurbineCard {...turbine} />
                    </Col>
                ))
                }
            </Row>
        </Container>
    );
};
