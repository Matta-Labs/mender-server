// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import React from 'react';

import { Code } from '@northern.tech/common-ui/copy-code';

const getDiffLineStyle = line => {
  if (line.startsWith('+ ')) {
    return 'green';
  } else if (line.startsWith('- ')) {
    return 'red';
  }
  return '';
};

export const UserChange = ({ item }) => (
  <Code className="flexbox column">
    {item.change.split('\n').map((line, index) => (
      <span key={`line-${index}`} className={getDiffLineStyle(line)}>
        {line}
      </span>
    ))}
  </Code>
);

export default UserChange;
