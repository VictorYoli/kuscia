#!/bin/bash
#
# Copyright 2023 Ant Group Co., Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

echo "test suite env: TEST_SUITE_RUN_ROOT_DIR=${TEST_SUITE_RUN_ROOT_DIR}"
echo "test suite env: TEST_BIN_DIR=${TEST_BIN_DIR}"

TEST_SUITE_P2P_TEST_RUN_DIR=${TEST_SUITE_RUN_ROOT_DIR}/p2p
TEST_SUITE_P2P_TEST_RUN_KUSCIA_DIR=${TEST_SUITE_P2P_TEST_RUN_DIR}/kusciaapi

mkdir -p "${TEST_SUITE_P2P_TEST_RUN_DIR}"

. ./test/suite/core/functions.sh

function oneTimeSetUp() {
  ALLOW_PRIVILEGED=true start_p2p_mode "${TEST_SUITE_P2P_TEST_RUN_KUSCIA_DIR}"
}

# Note: This will be call twice on master, because of a bug: https://github.com/kward/shunit2/issues/112. And it is open now.
# we have a bugfix on shunit2, line number: [1370-1371]
function oneTimeTearDown() {
  stop_p2p_mode
}

function test_p2p_kuscia_job() {
  local alice_job_id=secretflow-psi-p2p-alice-nsjail
  docker exec -it "${AUTONOMY_ALICE_CONTAINER}" scripts/user/create_example_job.sh NSJAIL_PSI ${alice_job_id}
  assertEquals "Kuscia job failed" "Succeeded" "$(wait_kuscia_job_until "${AUTONOMY_ALICE_CONTAINER}" 600 ${alice_job_id})"
  assertEquals "Kuscia data file exist" "Y" "$(exist_container_file "${AUTONOMY_ALICE_CONTAINER}" var/storage/data/psi-output.csv)"

  unset alice_job_id
}

. ./test/vendor/shunit2
