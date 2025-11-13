import { Fragment, type FC } from "react";
import { NavLink } from "react-router";

interface Breadcrumb {
    label: string;
    path: string;
}

interface BreadcrumbsProps {
    breadcrumbs: Breadcrumb[];
}

export const Breadcrumbs: FC<BreadcrumbsProps> = ({ breadcrumbs }) => (
    <span style={{ display: "flex", gap: 10 }}> {
        breadcrumbs.map((value, index) =>
            <Fragment key={index}>
                {index < breadcrumbs.length - 1
                    ? <>
                        <NavLink to={value.path} >
                            {value.label}
                        </NavLink >
                        {index < breadcrumbs.length - 1 ? "/" : ""}
                    </>
                    : <p>{value.label}</p>}
            </Fragment>
        )}
    </span>
)