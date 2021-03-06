#below lines are required by every tests to import sibling package
import sys
sys.path.append('..')
sys.path.append('../pb')

from framework import tester
from framework import logger
from framework import beaconnode
from framework import validatornode
from framework import util
from framework import rpc
from framework import context
from framework import asserts

from pb import rpc_pb2
from pb import common_pb2

class RpcGetValidatorInformation :
    def __init__(self) :
        logger.set_verbose(True)
        
    def run(self) :
        ctx = context.Context(
            #directory = '/temp/synapse',
            #delete_data_on_exit = False
        )
        self._tester = tester.Tester(ctx)
        self._tester.run(self._do_run)
        
    def _do_run(self) :
        beacon_config = util.decode_json(util.read_file('data/regtest.json'))
        beacon_config['NetworkID'] = 'testnet'
        beacon_config['GenesisTime'] = 1566275808
        beacon_node_list = self._tester.create_nodes(
            1,
            node_class = beaconnode.BeaconNode,
            node_config_list = [ beacon_config ]
        )
        
        self._tester.start_all_nodes()
        util.sleep_for_seconds(5)

        rpc_client = rpc.create_beacon_rpc(beacon_node_list[0].get_rpc_address())
        self.test_invalid_id(rpc_client)
        self.test_wrong_request_type(rpc_client)
        
    def test_invalid_id(self, rpc_client) :
        request = rpc_pb2.GetValidatorRequest()
        request.ID = 1000000
        try :
            response = rpc_client.GetValidatorInformation(request)
        except Exception as e :
            asserts.assert_exception_contain_text(e, "could not find validator with ID")

    def test_wrong_request_type(self, rpc_client) :
        logger.error('The RPC should return error here instead of accepting the request successfully.')

        request = rpc_pb2.MempoolRequest()
        request.LastBlockHash = util.make_random_hash()
        try :
            response = rpc_client.GetValidatorInformation(request)
        except Exception as e :
            asserts.assert_exception_contain_text(e, "could not find validator with ID")

RpcGetValidatorInformation().run()
