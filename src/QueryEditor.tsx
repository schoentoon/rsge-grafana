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
  const [blur, setBlur] = React.useState<boolean>(false);

  const [lastQuery, setLastQuery] = React.useState<LastQuery | null>(null);

  const [searchOptions, setSearchOptions] = React.useState<Array<SelectableValue<string>>>([]);
  const [isSearchOptionsLoading, setIsSearchOptionsLoading] = React.useState<boolean>(false);
  const [isInitial, setInitial] = React.useState<boolean>(true);

  if (isInitial) {
    // this is purely to load name and icon on initial load of an already existing panel
    if (query !== undefined && query.itemID !== undefined) {
      var id = parseInt(query.itemID, 10);
      if (!isNaN(id)) {
        datasource.idToItem(id).then(
          result => {
            if (result !== undefined) {
              setSearch({
                label: result.name,
                value: result.id.toString(),
                imgUrl: result.icon,
              });
              setBlur(true);
            }
          },
          response => {
            setSearch({ label: "Configured item doesn't exist??", value: '' });
            setSearchOptions([]);

            throw new Error(response.statusText);
          }
        );
      }
    }
  }

  const loadSearch = React.useCallback(() => {
    if (isInitial || blur) {
      setInitial(false);
      return undefined;
    }
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
  }, [datasource, search, query, blur, isInitial, setInitial]);

  const refreshSearchOptions = React.useCallback(() => {
    setIsSearchOptionsLoading(true);
    var searches = loadSearch();
    if (searches !== undefined) {
      searches
        .then(i => {
          setSearchOptions(i);
        })
        .finally(() => {
          setIsSearchOptionsLoading(false);
        });
    }
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
            disabled={blur}
            onChange={v => {
              setSearch(v);
            }}
          />
        </div>
      </div>
    </>
  );
};
