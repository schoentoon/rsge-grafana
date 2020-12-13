import { DataSourceInstanceSettings, SelectableValue } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';

class Item {
  id = 0;
  name = '';
  icon = '';
}

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  async searchItems(query: string | undefined): Promise<Array<SelectableValue<string>>> {
    if (query === undefined || query === '') {
      return [];
    }
    return this.postResource('searchItems', { query: query }).then((items: Item[]) => {
      return items
        ? Object.entries(items).map(
            ([_, item]) =>
              ({
                label: item.name,
                value: item.id.toString(),
                imgUrl: item.icon,
              } as SelectableValue<string>)
          )
        : [];
    });
  }
}
