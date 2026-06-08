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

import React, { useEffect, useState, useRef } from 'react';
import { Banner, Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

export default function SettingsCreditLimit(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    QuotaForNewUser: '',
    PreConsumedQuota: '',
    QuotaForInviter: '',
    QuotaForInvitee: '',
    AffiliateSignupRewardEnabled: false,
    AffiliateSignupRewardQuota: '',
    AffiliateIdentityEnabled: false,
    AffiliateIdentityConfig: '',
    AffiliateWithdrawEnabled: false,
    AffiliateWithdrawMinQuota: '',
    AffiliateWithdrawDailyLimit: '',
    AffiliateWithdrawHelpText: '',
    'quota_setting.enable_free_model_pre_consume': true,
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);
  const complianceConfirmed =
    props.options?.['payment_setting.compliance_confirmed'] === true ||
    props.options?.['payment_setting.compliance_confirmed'] === 'true';

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
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
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);
  return (
    <>
      <Spin spinning={loading}>
        {!complianceConfirmed && (
          <Banner
            type='warning'
            description={t(
              '设置非零邀请奖励额度前，需要先在支付设置中确认合规声明。',
            )}
            closeIcon={null}
            className='!rounded-lg mb-3'
          />
        )}
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('额度设置')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('新用户初始额度')}
                  field={'QuotaForNewUser'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  placeholder={''}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      QuotaForNewUser: String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('请求预扣费额度')}
                  field={'PreConsumedQuota'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  extraText={t('请求结束后多退少补')}
                  placeholder={''}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      PreConsumedQuota: String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('邀请新用户奖励额度')}
                  field={'QuotaForInviter'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  extraText={
                    !complianceConfirmed ? t('非零值需先确认合规声明') : ''
                  }
                  placeholder={t('例如：2000')}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      QuotaForInviter: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('新用户使用邀请码奖励额度')}
                  field={'QuotaForInvitee'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  extraText={
                    !complianceConfirmed ? t('非零值需先确认合规声明') : ''
                  }
                  placeholder={t('例如：1000')}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      QuotaForInvitee: String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.Switch
                  label={t('启用邀请注册奖励')}
                  field={'AffiliateSignupRewardEnabled'}
                  extraText={
                    !complianceConfirmed ? t('启用前需先确认合规声明') : ''
                  }
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateSignupRewardEnabled: value,
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('邀请注册奖励额度')}
                  field={'AffiliateSignupRewardQuota'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateSignupRewardQuota: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.Switch
                  label={t('启用邀请身份倍率')}
                  field={'AffiliateIdentityEnabled'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateIdentityEnabled: value,
                    })
                  }
                />
              </Col>
              <Col xs={24}>
                <Form.TextArea
                  label={t('邀请身份倍率配置')}
                  field={'AffiliateIdentityConfig'}
                  autosize
                  placeholder='{"inviter_rate_multiplier":1.5,"invitee_rate_multiplier":1.4,"duration_hours":720}'
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateIdentityConfig: value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.Switch
                  label={t('启用邀请返利提现')}
                  field={'AffiliateWithdrawEnabled'}
                  extraText={
                    !complianceConfirmed ? t('启用前需先确认合规声明') : ''
                  }
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateWithdrawEnabled: value,
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('邀请返利最低提现额度')}
                  field={'AffiliateWithdrawMinQuota'}
                  step={1}
                  min={0}
                  suffix={'Token'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateWithdrawMinQuota: String(value),
                    })
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={6}>
                <Form.InputNumber
                  label={t('每日提现申请次数')}
                  field={'AffiliateWithdrawDailyLimit'}
                  step={1}
                  min={0}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateWithdrawDailyLimit: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24}>
                <Form.TextArea
                  label={t('邀请返利提现说明')}
                  field={'AffiliateWithdrawHelpText'}
                  autosize
                  placeholder={t('例如：请填写微信收款账号或联系方式')}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      AffiliateWithdrawHelpText: value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col>
                <Form.Switch
                  label={t('对免费模型启用预消耗')}
                  field={'quota_setting.enable_free_model_pre_consume'}
                  extraText={t(
                    '开启后，对免费模型（倍率为0，或者价格为0）的模型也会预消耗额度',
                  )}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'quota_setting.enable_free_model_pre_consume': value,
                    })
                  }
                />
              </Col>
            </Row>

            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存额度设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
