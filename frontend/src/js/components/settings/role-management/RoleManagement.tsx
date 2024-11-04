// Copyright 2020 Northern.tech AS
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
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Add as AddIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';

import DetailsIndicator from '@northern.tech/common-ui/detailsindicator';
import DetailsTable from '@northern.tech/common-ui/detailstable';
import { DocsTooltip } from '@northern.tech/common-ui/docslink';
import EnterpriseNotification from '@northern.tech/common-ui/enterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/info-hint';
import { BENEFITS, emptyRole } from '@northern.tech/store/constants';
import { getGroupsByIdWithoutUngrouped, getIsEnterprise, getOrganization, getReleaseTagsById, getRelevantRoles } from '@northern.tech/store/selectors';
import { createRole, editRole, getDynamicGroups, getExistingReleaseTags, getGroups, getRoles, removeRole } from '@northern.tech/store/thunks';

import RoleDefinition from './RoleDefinition';

const columns = [
  { key: 'name', title: 'Role', render: ({ name }) => name },
  { key: 'description', title: 'Description', render: ({ description }) => description || '-' },
  { key: 'manage', title: 'Manage', render: DetailsIndicator }
];

export const RoleManagement = () => {
  const [adding, setAdding] = useState(false);
  const [editing, setEditing] = useState(false);
  const [role, setRole] = useState({ ...emptyRole });
  const dispatch = useDispatch();
  const groups = useSelector(getGroupsByIdWithoutUngrouped);
  const releaseTags = useSelector(getReleaseTagsById);
  const isEnterprise = useSelector(getIsEnterprise);
  const { service_provider } = useSelector(getOrganization);
  const items = useSelector(getRelevantRoles);

  useEffect(() => {
    dispatch(getExistingReleaseTags());
  }, [dispatch]);

  useEffect(() => {
    if (Object.keys(groups).length) {
      return;
    }
    dispatch(getDynamicGroups());
    dispatch(getGroups());
    dispatch(getRoles());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(groups)]);

  const addRole = () => {
    setAdding(true);
    setEditing(false);
    setRole({ ...emptyRole });
  };

  const onEditRole = editedRole => {
    setAdding(false);
    setEditing(true);
    setRole(editedRole);
  };

  const onCancel = () => {
    setAdding(false);
    setEditing(false);
  };

  const onSubmit = submittedRole => {
    if (adding) {
      dispatch(createRole(submittedRole));
    } else {
      dispatch(editRole(submittedRole));
    }
    onCancel();
  };

  return (
    <div>
      <div className="flexbox center-aligned">
        <h2 style={{ marginLeft: 20 }}>Roles</h2>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.rbac.id} />
          <DocsTooltip />
        </InfoHintContainer>
      </div>
      <DetailsTable columns={columns} items={items} onItemClick={onEditRole} />
      <Chip color="primary" icon={<AddIcon />} label="Add a role" onClick={addRole} disabled={!isEnterprise} />
      <RoleDefinition
        adding={adding}
        editing={editing}
        isServiceProvider={!!service_provider}
        onCancel={onCancel}
        onSubmit={onSubmit}
        removeRole={name => dispatch(removeRole(name))}
        selectedRole={role}
        stateGroups={groups}
        stateReleaseTags={releaseTags}
      />
    </div>
  );
};

export default RoleManagement;
