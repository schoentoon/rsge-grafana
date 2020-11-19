import defaults from "lodash/defaults";

import React, { ChangeEvent, PureComponent } from "react";
import { LegacyForms } from "@grafana/ui";
import { QueryEditorProps } from "@grafana/data";
import { DataSource } from "./DataSource";
import { defaultQuery, MyDataSourceOptions, MyQuery } from "./types";

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onItemIDChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, itemID: parseInt(event.target.value, 10) });
    // executes the query
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { itemID } = query;

    return (
      <div className="gf-form">
        <FormField
          width={4}
          value={itemID}
          onChange={this.onItemIDChange}
          label="Item ID"
          type="number"
        />
      </div>
    );
  }
}
