/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useRef, useState } from 'react';
import { Banner, Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  API,
  compareObjects,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

const smsDefaults = {
  PhoneVerificationEnabled: false,
  SMSIHuyiEnabled: false,
  SMSIHuyiAPIID: '',
  SMSIHuyiAPIKey: '',
  SMSIHuyiTemplateID: '',
};

export default function SettingsSMS(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState(smsDefaults);
  const [inputsRow, setInputsRow] = useState(smsDefaults);
  const refForm = useRef();

  function handleFieldChange(fieldName) {
    return (value) => {
      setInputs((current) => ({ ...current, [fieldName]: value }));
    };
  }

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) {
      showWarning(t('你似乎并没有修改什么'));
      return;
    }
    const requestQueue = updateArray.map((item) => {
      const value =
        typeof inputs[item.key] === 'boolean'
          ? String(inputs[item.key])
          : inputs[item.key];
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (res.includes(undefined)) {
          showError(t('部分保存失败，请重试'));
          return;
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = { ...smsDefaults };
    for (const key in props.options) {
      if (Object.keys(smsDefaults).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(currentInputs);
    if (refForm.current) {
      refForm.current.setValues(currentInputs);
    }
  }, [props.options]);

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => {
          refForm.current = formAPI;
        }}
        onSubmit={onSubmit}
      >
        <Banner
          type='info'
          description={t(
            '短信配置优先读取 .env，后台系统设置仅作为兜底配置。',
          )}
          closeIcon={null}
          className='mb-4'
        />
        <Row gutter={16}>
          <Col span={24}>
            <Form.Switch
              field='PhoneVerificationEnabled'
              label={t('手机号验证')}
              checked={inputs.PhoneVerificationEnabled}
              onChange={handleFieldChange('PhoneVerificationEnabled')}
            />
          </Col>
          <Col span={24}>
            <Form.Switch
              field='SMSIHuyiEnabled'
              label={t('启用互亿无线短信')}
              checked={inputs.SMSIHuyiEnabled}
              onChange={handleFieldChange('SMSIHuyiEnabled')}
            />
          </Col>
          <Col span={24}>
            <Form.Input
              field='SMSIHuyiAPIID'
              label={t('互亿无线 API ID')}
              placeholder={t('请输入互亿无线 API ID')}
              onChange={handleFieldChange('SMSIHuyiAPIID')}
            />
          </Col>
          <Col span={24}>
            <Form.Input
              field='SMSIHuyiAPIKey'
              label={t('互亿无线 API Key')}
              placeholder={t('留空则不修改已保存的密钥')}
              mode='password'
              onChange={handleFieldChange('SMSIHuyiAPIKey')}
            />
          </Col>
          <Col span={24}>
            <Form.Input
              field='SMSIHuyiTemplateID'
              label={t('互亿无线模板 ID')}
              placeholder={t('请输入互亿无线模板 ID')}
              onChange={handleFieldChange('SMSIHuyiTemplateID')}
            />
          </Col>
        </Row>
        <Button htmlType='submit' type='primary'>
          {t('保存')}
        </Button>
      </Form>
    </Spin>
  );
}
