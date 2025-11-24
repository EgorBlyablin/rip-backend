import { useEffect, useState, type FC, type FormEvent } from "react";
import { Button, Col, Container, Form, InputGroup, Row } from "react-bootstrap";
import { useDispatch } from "react-redux";
import { fetchTurbines } from "../api/turbines";
import SearchIcon from "../assets/search.svg?react";
import { Breadcrumbs } from "../components/breadcrumbs";
import { TurbineCard } from "../components/turbine-card";
import { ROUTE_LABELS, ROUTES } from "../routes";
import type { Turbine } from "../api/interfaces";
import { fetchRequestStats } from "../api/generation-requests";
import { useTurbinesFilter, setTurbinesFilter } from "../api/turbines";

import CalculatorIcon from "../assets/calculator.svg?react";

export const TurbinesListPage: FC = () => {
    const dispatch = useDispatch();

    const turbinesFilter = useTurbinesFilter();
    const [turbines, setTurbines] = useState<Turbine[]>([]);
    const [cartCount, setCartCount] = useState(0);

    useEffect(() => {
        fetchRequestStats().then(stats => stats && setCartCount(stats.TurbinesCount));
    }, []);

    useEffect(() => {
        const fetchTurbinesWrapper = async () => {
            const turbinesData = await fetchTurbines(turbinesFilter);
            setTurbines(turbinesData);
        };

        fetchTurbinesWrapper();
    }, [turbinesFilter]);

    // Handle form submission for search
    const handleSearchSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        dispatch(setTurbinesFilter((e.currentTarget.elements.namedItem("title-filter") as HTMLInputElement).value));
    };

    return (
        <Container className="pb-5">
            <Breadcrumbs
                breadcrumbs={[
                    { label: ROUTE_LABELS.HOME, path: ROUTES.HOME },
                    { label: ROUTE_LABELS.TURBINES_LIST, path: ROUTES.TURBINES_LIST },
                ]}
            />
            <Row
                style={{
                    display: "flex",
                    gap: 16,
                    justifyContent: "space-between"
                }}
                className="p-2 mb-1"
            >
                <Col md={"auto"} xs={12}>
                    <h1
                        style={{
                            textTransform: "uppercase",
                            fontWeight: "bold",
                            color: "#5B5B5B",
                            lineHeight: 1,
                            margin: 0
                        }}
                        >
                        Ветрогенераторы
                        {turbinesFilter && <> (поиск: "{turbinesFilter}")</>} {/* Use turbinesFilter here too */}
                    </h1>
                </Col>
                <Col md={"auto"} style={{ display: "flex", gap: 10, alignItems: "center" }}>
                    <Button
                        variant="secondary"
                        style={{ textWrap: "nowrap", fontSize: "1.5rem", height: "36px", marginRight: "0.5rem", display: "flex", alignItems: "center" }}
                    >
                        <CalculatorIcon width={"25px"} /> <span style={{
                            position: "absolute",
                            background: "white",
                            color: "black",
                            borderRadius: "100rem",
                            border: "1px solid black",
                            padding: "0.05rem 0.5rem",
                            lineHeight: 1,
                            fontSize: "0.7rem",
                            transform: "translate(100%, -100%)"
                        }}>{cartCount}</span>
                    </Button>
                    <Form onSubmit={handleSearchSubmit}>
                        <InputGroup>
                            <Form.Control
                                name="title-filter"
                                placeholder="Поиск"
                                defaultValue={turbinesFilter}
                                maxLength={20}
                            />
                            <Button type="submit" style={{ display: "flex", alignItems: "center", backgroundColor: "#5BA1D4", border: "none" }}>
                                <SearchIcon style={{ width: "1.2rem", height: "1.2rem" }} />
                            </Button>
                        </InputGroup>
                    </Form>
                </Col>
            </Row>
            <Row>
                {(turbines || []).map((turbine) => (
                    <Col key={turbine.id} className="p-2" xxl={3} lg={4} sm={6} xs={12}>
                        <TurbineCard {...turbine} />
                    </Col>
                ))
                }
            </Row>
        </Container>
    );
};
