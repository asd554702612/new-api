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
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  API,
  compareObjects,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

const promptGateDefaults = {
  'promptgate.enabled': false,
  'promptgate.base_url': '',
  'promptgate.api_key': '',
  'promptgate.input_enabled': true,
  'promptgate.output_enabled': true,
  'promptgate.stream_output_enabled': true,
  'promptgate.stream_fail_closed': true,
};

function normalizeBaseUrl(value) {
  return String(value || '')
    .trim()
    .replace(/\/+$/, '');
}

export default function SettingsPromptGate(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState(promptGateDefaults);
  const [inputsRow, setInputsRow] = useState(promptGateDefaults);
  const refForm = useRef();

  function handleFieldChange(fieldName) {
    return (value) => {
      setInputs((current) => ({ ...current, [fieldName]: value }));
    };
  }

  function onSubmit() {
    const normalizedInputs = {
      ...inputs,
      'promptgate.base_url': normalizeBaseUrl(inputs['promptgate.base_url']),
      'promptgate.api_key': String(inputs['promptgate.api_key'] || '').trim(),
    };
    const updateArray = compareObjects(normalizedInputs, inputsRow);
    if (!updateArray.length) {
      showWarning(t('你似乎并没有修改什么'));
      return;
    }
    const requestQueue = updateArray.map((item) =>
      API.put('/api/option/', {
        key: item.key,
        value:
          typeof normalizedInputs[item.key] === 'boolean'
            ? String(normalizedInputs[item.key])
            : normalizedInputs[item.key],
      }),
    );

    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (res.includes(undefined)) {
          showError(t('部分保存失败，请重试'));
          return;
        }
        showSuccess(t('保存成功'));
        setInputsRow(structuredClone(normalizedInputs));
        setInputs(normalizedInputs);
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
    const currentInputs = { ...promptGateDefaults };
    for (const key in props.options) {
      if (Object.keys(promptGateDefaults).includes(key)) {
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
        <Form.Section text={t('PromptGate 设置')}>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Switch
                field='promptgate.enabled'
                label={t('启用 PromptGate')}
                checked={inputs['promptgate.enabled']}
                onChange={handleFieldChange('promptgate.enabled')}
              />
            </Col>
            <Col span={24}>
              <Form.Input
                field='promptgate.base_url'
                label={t('PromptGate API 地址')}
                placeholder={t('http://127.0.0.1:8080')}
                onChange={handleFieldChange('promptgate.base_url')}
              />
            </Col>
            <Col span={24}>
              <Form.Input
                field='promptgate.api_key'
                label={t('PromptGate API Key')}
                placeholder={t('留空则不修改已保存的密钥')}
                mode='password'
                onChange={handleFieldChange('promptgate.api_key')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.Switch
                field='promptgate.input_enabled'
                label={t('输入审核')}
                checked={inputs['promptgate.input_enabled']}
                onChange={handleFieldChange('promptgate.input_enabled')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.Switch
                field='promptgate.output_enabled'
                label={t('输出审核')}
                checked={inputs['promptgate.output_enabled']}
                onChange={handleFieldChange('promptgate.output_enabled')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.Switch
                field='promptgate.stream_output_enabled'
                label={t('流式输出审核')}
                checked={inputs['promptgate.stream_output_enabled']}
                onChange={handleFieldChange('promptgate.stream_output_enabled')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.Switch
                field='promptgate.stream_fail_closed'
                label={t('流式失败关闭')}
                checked={inputs['promptgate.stream_fail_closed']}
                onChange={handleFieldChange('promptgate.stream_fail_closed')}
              />
            </Col>
          </Row>
          <Button htmlType='submit' type='primary'>
            {t('保存')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
