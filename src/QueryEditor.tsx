import React, { ComponentType } from 'react';
import { Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './DataSource';
import { MyDataSourceOptions, MyQuery } from './types';
import { find } from 'lodash';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

interface LastQuery {
  search: string;
}

export const QueryEditor: ComponentType<Props> = ({ datasource, onChange, onRunQuery, query }) => {
  const [search, setSearch] = React.useState<SelectableValue<string>>();

  const [lastQuery, setLastQuery] = React.useState<LastQuery | null>(null);

  const [searchOptions, setSearchOptions] = React.useState<Array<SelectableValue<string>>>([]);
  const [isSearchOptionsLoading, setIsSearchOptionsLoading] = React.useState<boolean>(false);

  const loadSearch = React.useCallback(() => {
    var q = '';
    if (search !== undefined && search.value !== undefined) {
      q = search.value;
    }
    return datasource.searchItems(q).then(
      result => {
        const searches = result;

        const foundSearch = find(searches, i => i.value === q);

        if (foundSearch !== undefined && foundSearch.value !== undefined) {
          var id = parseInt(foundSearch.value, 10);
          if (!isNaN(id)) {
            query.itemID = foundSearch.value;
          }
        }

        return searches;
      },
      response => {
        setSearch({ label: 'No search results :(', value: '' });
        setSearchOptions([]);

        throw new Error(response.statusText);
      }
    );
  }, [datasource, search, query]);

  const refreshSearchOptions = React.useCallback(() => {
    setIsSearchOptionsLoading(true);
    loadSearch()
      .then(i => {
        setSearchOptions(i);
      })
      .finally(() => {
        setIsSearchOptionsLoading(false);
      });
  }, [loadSearch, setSearchOptions, setIsSearchOptionsLoading]);

  // Initializing metric options
  React.useEffect(() => {
    refreshSearchOptions();
  }, [refreshSearchOptions]);

  React.useEffect(() => {
    if (search?.value === undefined) {
      return;
    }

    if (lastQuery !== null && search?.value === lastQuery.search) {
      return;
    }

    setLastQuery({
      search: search.value,
    });

    var id = parseInt(search.value, 10);
    if (!isNaN(id)) {
      query.itemID = search.value;

      onChange({ ...query });

      onRunQuery();
    }
  }, [search, query, lastQuery, onChange, onRunQuery]);

  return (
    <>
      <div className="gf-form-inline">
        <div className="gf-form">
          <Select
            isLoading={isSearchOptionsLoading}
            width={40}
            prefix="Item: "
            options={searchOptions}
            placeholder="Select item"
            allowCustomValue
            value={search}
            onChange={v => {
              setSearch(v);
            }}
          />
        </div>
      </div>
    </>
  );
};
